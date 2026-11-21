use crate::error::BoxError;
use std::io::{Cursor, Read};

/// Source 2 (CS2) demo file signature. demoparser2 itself rejects anything
/// else (it recognizes "HL2DEMO\0" too, but only to report the older
/// Source 1 format as unsupported) - we check for it upfront so a bad
/// upload fails fast with a clear message instead of deep inside parsing.
const DEM_MAGIC: &[u8] = b"PBDEMS2\0";
const ZSTD_MAGIC: &[u8] = &[0x28, 0xB5, 0x2F, 0xFD];
const ZIP_MAGIC: &[u8] = &[0x50, 0x4B, 0x03, 0x04];
const SEVENZ_MAGIC: &[u8] = &[0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C];

/// Accepts raw uploaded bytes. They may already be a .dem file, or a
/// .zst/.zip/.7z archive wrapping one (a common way to share demos, since
/// they're large). Returns the raw .dem bytes, or an error if the input is
/// neither a demo nor a supported archive containing one.
pub fn extract_demo_bytes(raw: Vec<u8>) -> Result<Vec<u8>, BoxError> {
    if starts_with(&raw, DEM_MAGIC) {
        return Ok(raw);
    }
    if starts_with(&raw, ZSTD_MAGIC) {
        return require_dem(zstd::stream::decode_all(Cursor::new(raw))?, "zst");
    }
    if starts_with(&raw, ZIP_MAGIC) {
        return require_dem(extract_first_dem_from_zip(&raw)?, "zip");
    }
    if starts_with(&raw, SEVENZ_MAGIC) {
        return require_dem(extract_first_dem_from_7z(&raw)?, "7z");
    }

    Err("unrecognized file: expected a .dem demo file, or a .zst/.zip/.7z archive containing one".into())
}

fn starts_with(data: &[u8], magic: &[u8]) -> bool {
    data.len() >= magic.len() && &data[..magic.len()] == magic
}

fn require_dem(data: Vec<u8>, archive_kind: &str) -> Result<Vec<u8>, BoxError> {
    if starts_with(&data, DEM_MAGIC) {
        Ok(data)
    } else {
        Err(format!("{archive_kind} archive did not contain a valid .dem file").into())
    }
}

fn extract_first_dem_from_zip(raw: &[u8]) -> Result<Vec<u8>, BoxError> {
    let mut archive = zip::ZipArchive::new(Cursor::new(raw))?;
    for i in 0..archive.len() {
        let mut entry = archive.by_index(i)?;
        let name = match entry.name() {
            Ok(name) => name.into_owned(),
            Err(_) => continue,
        };
        if name.ends_with(".dem") {
            let mut buf = Vec::new();
            entry.read_to_end(&mut buf)?;
            return Ok(buf);
        }
    }
    Err("zip archive does not contain a .dem entry".into())
}

fn extract_first_dem_from_7z(raw: &[u8]) -> Result<Vec<u8>, BoxError> {
    let temp_dir = tempfile::tempdir()?;
    sevenz_rust::decompress(Cursor::new(raw), temp_dir.path())
        .map_err(|e| format!("failed to decompress 7z archive: {e}"))?;

    for entry in std::fs::read_dir(temp_dir.path())? {
        let path = entry?.path();
        if path.extension().and_then(|ext| ext.to_str()) == Some("dem") {
            return Ok(std::fs::read(path)?);
        }
    }
    Err("7z archive does not contain a .dem entry".into())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;

    fn fake_dem(body: &[u8]) -> Vec<u8> {
        [DEM_MAGIC, body].concat()
    }

    #[test]
    fn passes_through_a_valid_dem_unchanged() {
        let dem = fake_dem(b"rest of the demo");
        assert_eq!(extract_demo_bytes(dem.clone()).unwrap(), dem);
    }

    #[test]
    fn rejects_unrecognized_input() {
        let err = extract_demo_bytes(b"not a demo or an archive".to_vec()).unwrap_err();
        assert!(err.to_string().contains("unrecognized file"));
    }

    #[test]
    fn decompresses_a_zst_wrapped_dem() {
        let dem = fake_dem(b"zst payload");
        let compressed = zstd::stream::encode_all(Cursor::new(&dem), 0).unwrap();
        assert_eq!(extract_demo_bytes(compressed).unwrap(), dem);
    }

    #[test]
    fn rejects_a_zst_archive_not_containing_a_dem() {
        let compressed = zstd::stream::encode_all(Cursor::new(b"not a demo".as_slice()), 0).unwrap();
        let err = extract_demo_bytes(compressed).unwrap_err();
        assert!(err.to_string().contains("zst archive"));
    }

    #[test]
    fn extracts_a_dem_from_a_zip_archive() {
        let dem = fake_dem(b"zip payload");
        let mut buf = Vec::new();
        {
            let mut writer = zip::ZipWriter::new(Cursor::new(&mut buf));
            writer.start_file("match.dem", zip::write::SimpleFileOptions::default()).unwrap();
            writer.write_all(&dem).unwrap();
            writer.finish().unwrap();
        }
        assert_eq!(extract_demo_bytes(buf).unwrap(), dem);
    }

    #[test]
    fn rejects_a_zip_archive_without_a_dem_entry() {
        let mut buf = Vec::new();
        {
            let mut writer = zip::ZipWriter::new(Cursor::new(&mut buf));
            writer.start_file("readme.txt", zip::write::SimpleFileOptions::default()).unwrap();
            writer.write_all(b"not a demo").unwrap();
            writer.finish().unwrap();
        }
        let err = extract_demo_bytes(buf).unwrap_err();
        assert!(err.to_string().contains("zip archive does not contain"));
    }

    #[test]
    fn extracts_a_dem_from_a_7z_archive() {
        let dem = fake_dem(b"7z payload");
        let src_dir = tempfile::tempdir().unwrap();
        let dem_path = src_dir.path().join("match.dem");
        std::fs::write(&dem_path, &dem).unwrap();

        let mut buf = Vec::new();
        sevenz_rust::compress(&dem_path, Cursor::new(&mut buf)).unwrap();

        assert_eq!(extract_demo_bytes(buf).unwrap(), dem);
    }
}
