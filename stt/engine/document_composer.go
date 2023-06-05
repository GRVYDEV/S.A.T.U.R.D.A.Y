package engine

type Document struct {
	TranscribedText      string
	NewText              string
	CurrentTranscription string
}

type DocumentComposer struct {
	transcriptions        []*Transcription
	finishedText          string
	finishedTextTimeStamp uint32
}

func NewDocumentComposer() *DocumentComposer {
	documentComposer := &DocumentComposer{
		// TODO maybe set max size?
		transcriptions: make([]*Transcription, 0),
	}

	return documentComposer
}

func (dc *DocumentComposer) NewTranscript(script Transcription) (Document, uint32) {
	dc.transcriptions = append(dc.transcriptions, &script)
	return dc.ComposeDocument()
}

func (dc *DocumentComposer) ComposeDocument() (Document, uint32) {
	var document Document

	transcriptions := DeepCopyTranscriptions(dc.transcriptions)
	flippiSegment := ""
	for {
		choosenTranscription := FindOldestTranscription(transcriptions)
		if choosenTranscription == nil {
			break
		}

		// Delete it and every intersecting transcription from the list. The only transcripts left will be the ones later in the timeline
		transcriptions = DeleteIntersectingTranscriptions(choosenTranscription, transcriptions)

		for i, segment := range choosenTranscription.Transcriptions {

			if i == len(choosenTranscription.Transcriptions)-1 {
				if len(transcriptions) == 0 {

					if flippiSegment != "" {
						flippiSegment += " "
					}
					flippiSegment += segment.Text
					break
				}
			}
			if document.NewText != "" {
				document.NewText += " "
			}
			document.NewText += segment.Text
			dc.finishedTextTimeStamp = choosenTranscription.From + segment.EndTimestamp

			Logger.Infof("choosenTranscription.From: %d", choosenTranscription.From)
			Logger.Infof("segment.EndTimestamp: %d", segment.EndTimestamp)
		}
	}

	dc.finishedText += document.NewText
	if document.TranscribedText != "" {
		document.TranscribedText += " "
	}
	document.TranscribedText = dc.finishedText
	document.CurrentTranscription = flippiSegment
	dc.DeleteSegmentsContaining(dc.finishedTextTimeStamp)
	return document, dc.finishedTextTimeStamp
}

func (dc *DocumentComposer) DeleteSegmentsContaining(timestamp uint32) {
	var updatedTranscriptions []*Transcription

	for _, transcription := range dc.transcriptions {
		var updatedSegments []TranscriptionSegment

		for _, segment := range transcription.Transcriptions {
			if timestamp > transcription.From+segment.StartTimestamp {
				continue // Skip the segment containing the specified timestamp
			}

			updatedSegments = append(updatedSegments, segment)
		}

		if len(updatedSegments) > 0 {
			updatedTranscriptions = append(updatedTranscriptions, &Transcription{
				From:           transcription.From,
				Transcriptions: updatedSegments,
			})
		}
	}

	dc.transcriptions = updatedTranscriptions
}

func FindOldestTranscription(transcriptions []*Transcription) *Transcription {
	var lastTranscription *Transcription
	latestTimeStamp := uint32(0)

	for _, transcription := range transcriptions {
		// if transcription has no segments then skip
		if len(transcription.Transcriptions) == 0 {
			continue
		}
		if lastTranscription == nil {
			lastTranscription = transcription
			latestTimeStamp = transcription.From + transcription.Transcriptions[0].StartTimestamp
		} else if transcription.From < latestTimeStamp {
			latestTimeStamp = transcription.From + transcription.Transcriptions[0].StartTimestamp
			lastTranscription = transcription
		} else if areWithin100(transcription.From, latestTimeStamp) && isTranscriptionLonger(transcription, lastTranscription) {
			lastTranscription = transcription
		}
	}

	return lastTranscription
}

func DeleteIntersectingTranscriptions(chosen *Transcription, transcriptions []*Transcription) []*Transcription {
	updatedTranscriptions := make([]*Transcription, 0)

	for _, transcription := range transcriptions {
		if transcription != chosen {
			newSegments := make([]TranscriptionSegment, 0)

			for _, segment := range transcription.Transcriptions {
				intersecting := false
				for _, chosenSegment := range chosen.Transcriptions {
					chosenStart := chosen.From + chosenSegment.StartTimestamp
					chosenEnd := chosen.From + chosenSegment.EndTimestamp
					segmentStart := transcription.From + segment.StartTimestamp
					segmentEnd := transcription.From + segment.EndTimestamp

					if (chosenStart < segmentEnd && chosenEnd > segmentStart) ||
						(segmentStart < chosenEnd && segmentEnd > chosenStart) {
						intersecting = true
						break
					}
				}
				if !intersecting {
					newSegments = append(newSegments, segment)
				}
			}

			if len(newSegments) > 0 {
				newTranscription := DeepCopyTranscription(transcription)
				newTranscription.Transcriptions = newSegments
				updatedTranscriptions = append(updatedTranscriptions, newTranscription)
			}
		}
	}

	return updatedTranscriptions
}

func isTranscriptionLonger(a *Transcription, b *Transcription) bool {
	return a.Transcriptions[len(a.Transcriptions)-1].EndTimestamp > b.Transcriptions[len(b.Transcriptions)-1].EndTimestamp
}

func areWithin100(a, b uint32) bool {
	if a > b {
		return a-b <= 100
	}
	return b-a <= 100
}

func DeepCopyTranscriptions(t []*Transcription) []*Transcription {
	copy := make([]*Transcription, len(t))
	for i, transcription := range t {
		copy[i] = DeepCopyTranscription(transcription)
	}
	return copy
}

func DeepCopyTranscription(t *Transcription) *Transcription {
	copy := Transcription{
		From:           t.From,
		Transcriptions: make([]TranscriptionSegment, len(t.Transcriptions)),
	}
	for i, segment := range t.Transcriptions {
		copy.Transcriptions[i] = TranscriptionSegment{
			StartTimestamp: segment.StartTimestamp,
			EndTimestamp:   segment.EndTimestamp,
			Text:           segment.Text,
		}
	}
	return &copy
}
