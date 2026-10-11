import QRCode from "qrcode";

/**
 * Standard ISO/IEC 18004 QR Code Generator.
 * Uses 'qrcode' under Level M error correction for reliable decodability.
 */

export function generateQrMatrix(text: string): boolean[][] {
  const qr = QRCode.create(text, { errorCorrectionLevel: "M" });
  const size = qr.modules.size;
  const data = qr.modules.data;
  const matrix: boolean[][] = [];
  for (let r = 0; r < size; r++) {
    const row: boolean[] = [];
    for (let c = 0; c < size; c++) {
      row.push(Boolean(data[r * size + c]));
    }
    matrix.push(row);
  }
  return matrix;
}

export function generateQrSvg(text: string, size = 160): string {
  const matrix = generateQrMatrix(text);
  const n = matrix.length;
  const padding = 2;
  const total = n + padding * 2;
  const cellSize = size / total;

  let rects = "";
  for (let r = 0; r < n; r++) {
    for (let c = 0; c < n; c++) {
      if (matrix[r][c]) {
        const x = (c + padding) * cellSize;
        const y = (r + padding) * cellSize;
        rects += `<rect x="${x.toFixed(2)}" y="${y.toFixed(2)}" width="${cellSize.toFixed(2)}" height="${cellSize.toFixed(2)}" fill="#111827"/>`;
      }
    }
  }

  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}" width="${size}" height="${size}" shape-rendering="crispEdges" data-qr-size="${n}"><rect width="${size}" height="${size}" fill="#ffffff"/>${rects}</svg>`;
}

export function drawQrToCanvas(
  canvas: HTMLCanvasElement,
  text: string,
  targetSize = 160,
) {
  const matrix = generateQrMatrix(text);
  const n = matrix.length;
  const padding = 2;
  const total = n + padding * 2;
  canvas.width = targetSize;
  canvas.height = targetSize;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  ctx.fillStyle = "#ffffff";
  ctx.fillRect(0, 0, targetSize, targetSize);
  ctx.fillStyle = "#111827";

  const cellSize = targetSize / total;
  for (let r = 0; r < n; r++) {
    for (let c = 0; c < n; c++) {
      if (matrix[r][c]) {
        ctx.fillRect(
          Math.floor((c + padding) * cellSize),
          Math.floor((r + padding) * cellSize),
          Math.ceil(cellSize),
          Math.ceil(cellSize),
        );
      }
    }
  }
}
