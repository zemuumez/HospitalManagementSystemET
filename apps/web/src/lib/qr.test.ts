import test from "node:test";
import assert from "node:assert/strict";
import jsQR from "jsqr";
import { generateQrMatrix, generateQrSvg } from "./qr.ts";

test("generateQrMatrix generates matrix decodable by jsQR", () => {
  const sampleUrl =
    "http://127.0.0.1:3000/v1/smart-cards/verify?token=sc-tok-1234567890abcdef1234567890abcdef";
  const matrix = generateQrMatrix(sampleUrl);
  assert.ok(Array.isArray(matrix));
  assert.ok(matrix.length > 20);

  const n = matrix.length;
  const padding = 2;
  const total = n + padding * 2;
  const targetSize = 200;
  const cellSize = targetSize / total;

  const imgData = new Uint8ClampedArray(targetSize * targetSize * 4);
  imgData.fill(255);

  for (let r = 0; r < n; r++) {
    for (let c = 0; c < n; c++) {
      if (matrix[r][c]) {
        const rx = Math.floor((c + padding) * cellSize);
        const ry = Math.floor((r + padding) * cellSize);
        const rw = Math.ceil(cellSize);
        const rh = Math.ceil(cellSize);
        for (let y = ry; y < ry + rh && y < targetSize; y++) {
          for (let x = rx; x < rx + rw && x < targetSize; x++) {
            const idx = (y * targetSize + x) * 4;
            imgData[idx] = 17;
            imgData[idx + 1] = 24;
            imgData[idx + 2] = 39;
          }
        }
      }
    }
  }

  const result = jsQR(imgData, targetSize, targetSize);
  assert.ok(result, "jsQR must successfully decode the rendered QR code");
  assert.equal(
    result.data,
    sampleUrl,
    "Decoded QR data must match original input text",
  );
});

test("generateQrSvg produces valid SVG representation", () => {
  const text = "http://127.0.0.1:3000/v1/smart-cards/verify?token=test-token";
  const svg = generateQrSvg(text, 160);
  assert.ok(svg.startsWith("<svg"));
  assert.ok(svg.endsWith("</svg>"));
  assert.ok(svg.includes('viewBox="0 0 160 160"'));
  assert.ok(svg.includes("<rect"));
});
