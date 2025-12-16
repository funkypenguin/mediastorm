#!/usr/bin/env python3
"""
Python wrapper for PTT (parsett) parse_title function.
Accepts a title string as a command-line argument and outputs JSON.
"""
import sys
import json
import PTT


def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No title provided"}), file=sys.stderr)
        sys.exit(1)

    title = sys.argv[1]

    try:
        result = PTT.parse_title(title)
        # Convert any non-serializable objects to strings
        print(json.dumps(result, default=str))
    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
