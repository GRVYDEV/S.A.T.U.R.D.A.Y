package engine

import (
	"testing"
)

func TestSimpleDocumentComposer(t *testing.T) {

	documentComposer := NewDocumentComposer()

	document, _ := documentComposer.NewTranscript(Transcription{
		Transcriptions: []TranscriptionSegment{
			{
				Text:           "Hello",
				StartTimestamp: 0,
				EndTimestamp:   20,
			},
		},
	})

	if document.CurrentTranscription != "Hello" {
		t.Errorf("Expected document text to be 'Hello', got %s", document.CurrentTranscription)
	}
}

func TestChangingWordTestDocumentComposer(t *testing.T) {
	documentComposer := NewDocumentComposer()

	document, _ := documentComposer.NewTranscript(Transcription{
		Transcriptions: []TranscriptionSegment{
			{
				Text:           "Hello",
				StartTimestamp: 0,
				EndTimestamp:   20,
			},
		},
	})
	if document.CurrentTranscription != "Hello" {
		t.Errorf("Expected document.CurrentTranscription to be 'Hello', got %s", document.CurrentTranscription)
	}
	if document.TranscribedText != "" {
		t.Errorf("Expected document.TranscribedText to be '', got %s", document.TranscribedText)
	}

	document, _ = documentComposer.NewTranscript(Transcription{
		Transcriptions: []TranscriptionSegment{
			{
				Text:           "Hello World",
				StartTimestamp: 0,
				EndTimestamp:   30,
			},
		},
	})
	if document.CurrentTranscription != "Hello World" {
		t.Errorf("Expected document.CurrentTranscription to be 'Hello World', got %s", document.CurrentTranscription)
	}
	if document.TranscribedText != "" {
		t.Errorf("Expected document.TranscribedText to be '', got %s", document.TranscribedText)
	}

	document, _ = documentComposer.NewTranscript(Transcription{
		Transcriptions: []TranscriptionSegment{
			{
				Text:           "Hello World",
				StartTimestamp: 0,
				EndTimestamp:   30,
			},
			{
				Text:           "this is a test",
				StartTimestamp: 30,
				EndTimestamp:   40,
			},
		},
	})
	if document.TranscribedText != "Hello World" {
		t.Errorf("Expected document.TranscribedText to be 'Hello World.', got %s", document.TranscribedText)
	}
	if document.CurrentTranscription != "this is a test" {
		t.Errorf("Expected document.CurrentTranscription to be 'this is a test', got %s", document.CurrentTranscription)
	}
}

func TestAddingTextToFinishedPreditction(t *testing.T) {
	documentComposer := NewDocumentComposer()

	document, _ := documentComposer.NewTranscript(Transcription{
		Transcriptions: []TranscriptionSegment{
			{
				Text:           "Hello",
				StartTimestamp: 0,
				EndTimestamp:   20,
			},
		},
	})

	if document.CurrentTranscription != "Hello" {
		t.Errorf("Expected document text to be 'Hello', got %s", document.CurrentTranscription)
	}

	document, _ = documentComposer.NewTranscript(Transcription{
		Transcriptions: []TranscriptionSegment{
			{
				Text:           "Hello",
				StartTimestamp: 0,
				EndTimestamp:   20,
			},
			{
				Text:           "this is a test",
				StartTimestamp: 20,
				EndTimestamp:   40,
			},
		},
	})
	if document.CurrentTranscription != "this is a test" {
		t.Errorf("Expected document text to be 'this is a test', got %s", document.CurrentTranscription)
	}
	if document.TranscribedText != "Hello" {
		t.Errorf("Expected document.TranscribedText to be 'Hello.', got %s", document.TranscribedText)
	}

	document, _ = documentComposer.NewTranscript(Transcription{
		Transcriptions: []TranscriptionSegment{
			{
				Text:           "this is a test.",
				StartTimestamp: 20,
				EndTimestamp:   40,
			},
			{
				Text:           "How is your day?",
				StartTimestamp: 40,
				EndTimestamp:   60,
			},
		},
	})
	if document.TranscribedText != "Hellothis is a test." {
		t.Errorf("Expected document.TranscribedText to be 'this is a test.', got %s", document.TranscribedText)
	}
	if document.CurrentTranscription != "How is your day?" {
		t.Errorf("Expected document text to be 'How is your day?', got %s", document.CurrentTranscription)
	}
}

func TestHavingNoTwoSegementsWithinOneWindow(t *testing.T) {
	documentComposer := NewDocumentComposer()

	document, _ := documentComposer.NewTranscript(Transcription{
		From: 100,
		Transcriptions: []TranscriptionSegment{
			{
				Text:           "Hello",
				StartTimestamp: 0,
				EndTimestamp:   700,
			},
		},
	})

	if document.CurrentTranscription != "Hello" {
		t.Errorf("Expected document text to be 'Hello', got %s", document.CurrentTranscription)
	}

	document, _ = documentComposer.NewTranscript(Transcription{
		From: 800,
		Transcriptions: []TranscriptionSegment{
			{
				Text:           "this is a test.",
				StartTimestamp: 0,
				EndTimestamp:   1200,
			},
			{
				Text:           "How is your day?",
				StartTimestamp: 1200,
				EndTimestamp:   2000,
			},
		},
	})
	if document.TranscribedText != "Hellothis is a test." {
		t.Errorf("Expected document.TranscribedText to be 'this is a test.', got %s", document.TranscribedText)
	}
	if document.CurrentTranscription != "How is your day?" {
		t.Errorf("Expected document text to be 'How is your day?', got %s", document.CurrentTranscription)
	}
}
