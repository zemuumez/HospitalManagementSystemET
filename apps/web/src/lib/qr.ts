/**
 * Self-contained, zero-dependency ISO/IEC 18004 standard QR Code Generator.
 * Supports Byte Mode encoding (8-bit bytes) with Reed-Solomon Error Correction.
 */

// Galois Field GF(2^8) math for QR error correction with primitive polynomial 0x11d (285)
const EXP_TABLE = new Uint8Array(512);
const LOG_TABLE = new Uint8Array(256);

(() => {
  let x = 1;
  for (let i = 0; i < 255; i++) {
    EXP_TABLE[i] = x;
    EXP_TABLE[i + 255] = x;
    LOG_TABLE[x] = i;
    x <<= 1;
    if (x & 256) x ^= 0x11d;
  }
})();

function gfMul(x: number, y: number): number {
  if (x === 0 || y === 0) return 0;
  return EXP_TABLE[LOG_TABLE[x] + LOG_TABLE[y]];
}

function rsGeneratorPoly(degree: number): Uint8Array {
  let poly = new Uint8Array([1]);
  for (let i = 0; i < degree; i++) {
    const next = new Uint8Array(poly.length + 1);
    const root = EXP_TABLE[i];
    for (let j = 0; j < poly.length; j++) {
      next[j] ^= gfMul(poly[j], root);
      next[j + 1] ^= poly[j];
    }
    poly = next;
  }
  return poly;
}

function rsComputeECC(data: Uint8Array, eccCount: number): Uint8Array {
  const gen = rsGeneratorPoly(eccCount);
  const remainder = new Uint8Array(eccCount);
  for (let i = 0; i < data.length; i++) {
    const factor = data[i] ^ remainder[0];
    remainder.copyWithin(0, 1);
    remainder[eccCount - 1] = 0;
    for (let j = 0; j < eccCount; j++) {
      remainder[j] ^= gfMul(gen[j], factor);
    }
  }
  return remainder;
}

// Version table for Level M (Medium) error correction
// [version, totalModules, dataCodewords, eccCodewordsPerBlock, numBlocks]
const VERSION_SPECS: [number, number, number, number, number][] = [
  [1, 21, 16, 10, 1],
  [2, 25, 28, 16, 1],
  [3, 29, 44, 26, 1],
  [4, 33, 64, 18, 2],
  [5, 37, 86, 24, 2],
  [6, 41, 108, 16, 4],
];

// Alignment pattern positions by version
const ALIGNMENT_COORDS: Record<number, number[]> = {
  2: [6, 18],
  3: [6, 22],
  4: [6, 26],
  5: [6, 30],
  6: [6, 34],
};

// Format info bit strings for Mask 000 (Level M: 00)
// Format bits 00000 -> 101010000010010 after BCH & XOR with 101010000010010 = 0
const FORMAT_INFO_M_MASK0 = [
  true,
  false,
  true,
  false,
  true,
  false,
  false,
  false,
  false,
  false,
  true,
  false,
  false,
  true,
  false,
];

export function generateQrMatrix(text: string): boolean[][] {
  const utf8 = new TextEncoder().encode(text);

  // Pick smallest fitting version
  let spec = VERSION_SPECS[0];
  for (const s of VERSION_SPECS) {
    if (s[2] - 2 >= utf8.length) {
      spec = s;
      break;
    }
    spec = s;
  }

  const [version, size, totalDataBytes, eccPerBlock, numBlocks] = spec;

  // Build bit stream: 4 bits mode (0100 byte), 8 bits length (for ver 1-9), data, 4 bits terminator
  const bits: number[] = [];
  function pushBits(val: number, len: number) {
    for (let i = len - 1; i >= 0; i--) {
      bits.push((val >> i) & 1);
    }
  }

  pushBits(0b0100, 4); // Byte mode
  pushBits(utf8.length, 8); // Character count indicator
  for (const b of utf8) {
    pushBits(b, 8);
  }

  // Terminator (up to 4 zeroes)
  const maxBits = totalDataBytes * 8;
  const termLen = Math.min(4, maxBits - bits.length);
  pushBits(0, termLen);

  // Pad to byte boundary
  while (bits.length % 8 !== 0) {
    bits.push(0);
  }

  // Pad with alternating 0xEC, 0x11
  const padBytes = [0xec, 0x11];
  let padIdx = 0;
  while (bits.length < maxBits) {
    pushBits(padBytes[padIdx % 2], 8);
    padIdx++;
  }

  // Convert bits to byte array
  const dataBytes = new Uint8Array(totalDataBytes);
  for (let i = 0; i < totalDataBytes; i++) {
    let b = 0;
    for (let j = 0; j < 8; j++) {
      b = (b << 1) | bits[i * 8 + j];
    }
    dataBytes[i] = b;
  }

  // Interleave blocks and compute Reed-Solomon ECC
  const blockSize = Math.floor(totalDataBytes / numBlocks);
  const dataBlocks: Uint8Array[] = [];
  const eccBlocks: Uint8Array[] = [];

  for (let b = 0; b < numBlocks; b++) {
    const start = b * blockSize;
    const end = b === numBlocks - 1 ? totalDataBytes : (b + 1) * blockSize;
    const blockData = dataBytes.slice(start, end);
    dataBlocks.push(blockData);
    eccBlocks.push(rsComputeECC(blockData, eccPerBlock));
  }

  // Interleave codewords
  const finalCodewords: number[] = [];
  const maxDataBlockLen = Math.max(...dataBlocks.map((b) => b.length));
  for (let i = 0; i < maxDataBlockLen; i++) {
    for (let b = 0; b < numBlocks; b++) {
      if (i < dataBlocks[b].length) {
        finalCodewords.push(dataBlocks[b][i]);
      }
    }
  }
  for (let i = 0; i < eccPerBlock; i++) {
    for (let b = 0; b < numBlocks; b++) {
      finalCodewords.push(eccBlocks[b][i]);
    }
  }

  // Initialize matrix
  const matrix: (boolean | null)[][] = Array.from({ length: size }, () =>
    Array(size).fill(null),
  );

  // Helper to place functional patterns
  function setFunc(r: number, c: number, v: boolean) {
    matrix[r][c] = v;
  }

  // Place Finder Pattern (7x7) + Separator
  function placeFinder(row: number, col: number) {
    for (let r = -1; r <= 7; r++) {
      for (let c = -1; c <= 7; c++) {
        const mr = row + r;
        const mc = col + c;
        if (mr >= 0 && mr < size && mc >= 0 && mc < size) {
          const isBlack =
            (r >= 0 && r <= 6 && (c === 0 || c === 6)) ||
            (c >= 0 && c <= 6 && (r === 0 || r === 6)) ||
            (r >= 2 && r <= 4 && c >= 2 && c <= 4);
          setFunc(mr, mc, isBlack);
        }
      }
    }
  }

  placeFinder(0, 0); // Top-Left
  placeFinder(0, size - 7); // Top-Right
  placeFinder(size - 7, 0); // Bottom-Left

  // Alignment patterns
  if (version >= 2 && ALIGNMENT_COORDS[version]) {
    const coords = ALIGNMENT_COORDS[version];
    for (const r of coords) {
      for (const c of coords) {
        // Skip finders
        if (
          (r === 6 && c === 6) ||
          (r === 6 && c === size - 7) ||
          (r === size - 7 && c === 6)
        )
          continue;
        for (let dr = -2; dr <= 2; dr++) {
          for (let dc = -2; dc <= 2; dc++) {
            const isBlack =
              Math.max(Math.abs(dr), Math.abs(dc)) === 2 ||
              (dr === 0 && dc === 0);
            setFunc(r + dr, c + dc, isBlack);
          }
        }
      }
    }
  }

  // Timing patterns
  for (let i = 8; i < size - 8; i++) {
    if (matrix[6][i] === null) setFunc(6, i, i % 2 === 0);
    if (matrix[i][6] === null) setFunc(i, 6, i % 2 === 0);
  }

  // Dark module
  setFunc(4 * version + 9, 8, true);

  // Format info area reservation
  for (let i = 0; i <= 8; i++) {
    if (matrix[8][i] === null) setFunc(8, i, false);
    if (matrix[i][8] === null) setFunc(i, 8, false);
  }
  for (let i = 0; i < 8; i++) {
    if (matrix[8][size - 1 - i] === null) setFunc(8, size - 1 - i, false);
    if (matrix[size - 1 - i][8] === null) setFunc(size - 1 - i, 8, false);
  }

  // Place codewords with Mask 0 ((row + col) % 2 === 0)
  const codewordBits: boolean[] = [];
  for (const b of finalCodewords) {
    for (let i = 7; i >= 0; i--) {
      codewordBits.push(((b >> i) & 1) === 1);
    }
  }

  let bitIdx = 0;
  let upward = true;
  for (let rightCol = size - 1; rightCol > 0; rightCol -= 2) {
    if (rightCol === 6) rightCol--; // Skip vertical timing column
    const cols = [rightCol, rightCol - 1];
    const rowRange = upward
      ? Array.from({ length: size }, (_, r) => size - 1 - r)
      : Array.from({ length: size }, (_, r) => r);

    for (const r of rowRange) {
      for (const c of cols) {
        if (matrix[r][c] === null) {
          const bit =
            bitIdx < codewordBits.length ? codewordBits[bitIdx] : false;
          bitIdx++;
          // Apply mask 0: (r + c) % 2 === 0
          const mask = (r + c) % 2 === 0;
          matrix[r][c] = bit !== mask;
        }
      }
    }
    upward = !upward;
  }

  // Write Format Information (Mask 0, Level M)
  const fi = FORMAT_INFO_M_MASK0;
  // Top-left
  setFunc(8, 0, fi[0]);
  setFunc(8, 1, fi[1]);
  setFunc(8, 2, fi[2]);
  setFunc(8, 3, fi[3]);
  setFunc(8, 4, fi[4]);
  setFunc(8, 5, fi[5]);
  setFunc(8, 7, fi[6]);
  setFunc(8, 8, fi[7]);
  setFunc(7, 8, fi[8]);
  setFunc(5, 8, fi[9]);
  setFunc(4, 8, fi[10]);
  setFunc(3, 8, fi[11]);
  setFunc(2, 8, fi[12]);
  setFunc(1, 8, fi[13]);
  setFunc(0, 8, fi[14]);

  // Bottom-left / Top-right
  for (let i = 0; i < 7; i++) {
    setFunc(size - 1 - i, 8, fi[i]);
  }
  for (let i = 7; i < 15; i++) {
    setFunc(8, size - 15 + i, fi[i]);
  }

  return matrix.map((row) => row.map((cell) => cell ?? false));
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
