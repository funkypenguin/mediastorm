#!/usr/bin/env python3
"""
Batch Python wrapper for PTT (parsett) parse_title function.
Accepts multiple titles via stdin (one per line) and outputs JSON array.
"""
import sys
import json
import PTT


def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No titles provided"}), file=sys.stderr)
        sys.exit(1)

    # Read all titles from command line arguments (skip script name)
    titles = sys.argv[1:]

    results = []
    for title in titles:
        try:
            parsed = PTT.parse_title(title)
            results.append({
                "title": title,
                "parsed": parsed,
                "error": None
            })
        except Exception as e:
            results.append({
                "title": title,
                "parsed": None,
                "error": str(e)
            })

    # Output as JSON array
    print(json.dumps(results, default=str))


if __name__ == "__main__":
    main()
