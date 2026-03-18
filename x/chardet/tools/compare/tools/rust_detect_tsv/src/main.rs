use chardetng::EncodingDetector;
use std::env;
use std::fs::File;
use std::io::{BufRead, BufReader, BufWriter, Write};

fn hex_decode(s: &str) -> Result<Vec<u8>, String> {
    let bytes = s.as_bytes();
    if bytes.len() % 2 != 0 {
        return Err("odd hex length".to_string());
    }
    let mut out = Vec::with_capacity(bytes.len() / 2);
    let mut i = 0;
    while i < bytes.len() {
        let h = from_hex(bytes[i]).ok_or_else(|| "bad hex".to_string())?;
        let l = from_hex(bytes[i + 1]).ok_or_else(|| "bad hex".to_string())?;
        out.push((h << 4) | l);
        i += 2;
    }
    Ok(out)
}

fn from_hex(c: u8) -> Option<u8> {
    match c {
        b'0'..=b'9' => Some(c - b'0'),
        b'a'..=b'f' => Some(c - b'a' + 10),
        b'A'..=b'F' => Some(c - b'A' + 10),
        _ => None,
    }
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<String> = env::args().collect();
    if args.len() != 3 {
        eprintln!("usage: detect_tsv <input.tsv> <output.tsv>");
        std::process::exit(2);
    }

    let input = BufReader::new(File::open(&args[1])?);
    let mut output = BufWriter::new(File::create(&args[2])?);

    for (lineno, line_res) in input.lines().enumerate() {
        let line = line_res?;
        if line.trim().is_empty() {
            continue;
        }
        if lineno == 0 && line.starts_with("id\t") {
            continue;
        }
        let parts: Vec<&str> = line.split('\t').collect();
        if parts.len() != 5 {
            continue;
        }
        let id = parts[0];
        let tld_raw = parts[2];
        let allow_utf8 = parts[3] == "1";
        let bytes = match hex_decode(parts[4]) {
            Ok(v) => v,
            Err(_) => continue,
        };

        let mut detector = EncodingDetector::new();
        detector.feed(&bytes, true);
        let tld_opt = if tld_raw.is_empty() { None } else { Some(tld_raw.as_bytes()) };
        let enc = detector.guess(tld_opt, allow_utf8);
        writeln!(output, "{}\t{}", id, enc.name())?;
    }

    output.flush()?;
    Ok(())
}
