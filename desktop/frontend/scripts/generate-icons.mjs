import { mkdir, readFile, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import sharp from "sharp";

// All platform icons are rendered from the same editable vector.
const source = await readFile(new URL("../public/net-test.svg", import.meta.url));
const output = new URL("../../build/icons/", import.meta.url);
await mkdir(output, { recursive: true });
const sizes = [16, 24, 32, 48, 64, 128, 256, 512, 1024];
const images = new Map(await Promise.all(sizes.map(async size => [
  size,
  await sharp(source, { density: 384 }).resize(size, size).png().toBuffer(),
])));

await writeFile(new URL("net-test.png", output), images.get(512));
await writeFile(new URL("../../../Icon.png", import.meta.url), images.get(1024));

// ICO directory entries point to full PNG images (Windows Vista and later).
const windowsSizes = sizes.filter(size => size <= 256);
const directory = Buffer.alloc(6 + windowsSizes.length * 16);
directory.writeUInt16LE(1, 2);
directory.writeUInt16LE(windowsSizes.length, 4);
let offset = directory.length;
windowsSizes.forEach((size, index) => {
  const entry = 6 + index * 16;
  const png = images.get(size);
  directory[entry] = directory[entry + 1] = size === 256 ? 0 : size;
  directory.writeUInt16LE(1, entry + 4);
  directory.writeUInt16LE(32, entry + 6);
  directory.writeUInt32LE(png.length, entry + 8);
  directory.writeUInt32LE(offset, entry + 12);
  offset += png.length;
});
await writeFile(new URL("net-test.ico", output), Buffer.concat([
  directory, ...windowsSizes.map(size => images.get(size)),
]));

// ICNS uses big-endian chunks; Retina entries retain their own logical size.
const iconTypes = {
  icp4: 16, icp5: 32, icp6: 64, ic07: 128, ic08: 256, ic09: 512,
  ic10: 1024, ic11: 32, ic12: 64, ic13: 256, ic14: 512,
};
const chunks = Object.entries(iconTypes).map(([type, size]) => {
  const png = images.get(size);
  const header = Buffer.alloc(8);
  header.write(type);
  header.writeUInt32BE(png.length + 8, 4);
  return Buffer.concat([header, png]);
});
const header = Buffer.alloc(8);
header.write("icns");
header.writeUInt32BE(8 + chunks.reduce((sum, chunk) => sum + chunk.length, 0), 4);
await writeFile(new URL("net-test.icns", output), Buffer.concat([header, ...chunks]));
console.log(`Generated PNG, ICO, and ICNS icons in ${fileURLToPath(output)}`);
