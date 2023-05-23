const puppeteer = require("puppeteer");
const express = require("express");

// const app = express();

// app.use(express.static("web"));

// app.listen(8080, () => {});

(async () => {
  let roomName = process.env.ROOM || "test";

  let testerMinutes = parseInt(process.env.DURATION);
  if (!testerMinutes) {
    testerMinutes = 30;
  }

  let tabCount = parseInt(process.env.TABS);
  if (!tabCount) {
    tabCount = 1;
  }

  const browser = await puppeteer.launch({
    headless: true,
    dumpio: true,
    args: [
      "--disable-gpu",
      "--no-sandbox",
      "--use-gl=swiftshader",
      "--disable-dev-shm-usage",
      "--use-fake-ui-for-media-stream",
      "--use-fake-device-for-media-stream",
      "--use-file-for-fake-audio-capture=audio.wav",
      "--autoplay-policy=no-user-gesture-required",
      "--unsafely-treat-insecure-origin-as-secure=http:docker.internal:8080",
    ],
    ignoreDefaultArgs: ["--mute-audio"],
  });

  for (var i = 0; i < tabCount; i++) {
    const url = `http://localhost:8088?room=${roomName}&noSub=true`;
    const page = await browser.newPage();
    page
      .on("console", (message) =>
        console.log(
          `${message.type().substr(0, 3).toUpperCase()} ${message.text()}`
        )
      )
      .on("pageerror", ({ message }) => console.log(message))
      .on("response", (response) =>
        console.log(`${response.status()} ${response.url()}`)
      )
      .on("requestfailed", (request) =>
        console.log(`${request.failure().errorText} ${request.url()}`)
      );

    await page.setViewport({
      width: 1000,
      height: 700,
    });
    await page.goto(url, { waitUntil: "load" });
  }

  await sleep((testerMinutes + Math.random()) * 60 * 1000);
  await browser.close();
})();

function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}
