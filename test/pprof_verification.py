#!/usr/bin/env python3
"""Verify pytorch_to_pprof conversion by comparing input JSON and output pprof statistics."""

import json
import subprocess
import sys
import re

def analyze_input_json(json_path):
    with open(json_path) as f:
        data = json.load(f)

    events = [e for e in data['traceEvents'] if e.get('ph') == 'X' and e.get('dur', 0) > 0]
    total_dur_ns = sum(e['dur'] * 1000 for e in events)  # us to ns

    print(f'Input JSON: {json_path}')
    print(f'  Total X events with dur > 0: {len(events)}')
    print(f'  Total duration (sum of all): {total_dur_ns:,.0f} ns')
    print(f'  Total duration: {total_dur_ns / 1e9:.3f} seconds')

    return len(events), total_dur_ns

def analyze_output_pprof(pprof_path):
    result = subprocess.run(
        ['go', 'tool', 'pprof', '-top', '-unit=nanoseconds', pprof_path],
        capture_output=True, text=True
    )
    output = result.stdout

    # Parse total from "X% of Y total" line
    match = re.search(r'of (\d+)ns total', output)
    total_ns = int(match.group(1)) if match else 0

    print(f'\nOutput pprof: {pprof_path}')
    print(f'  Total duration: {total_ns:,.0f} ns')
    print(f'  Total duration: {total_ns / 1e9:.3f} seconds')

    return total_ns

def main():
    if len(sys.argv) != 3:
        print(f'Usage: {sys.argv[0]} <input.json> <output.pb.gz>')
        sys.exit(1)

    json_path = sys.argv[1]
    pprof_path = sys.argv[2]

    event_count, json_total = analyze_input_json(json_path)
    pprof_total = analyze_output_pprof(pprof_path)

    diff = abs(json_total - pprof_total)
    diff_pct = (diff / json_total * 100) if json_total > 0 else 0

    print(f'\nComparison:')
    print(f'  Difference: {diff:,.0f} ns ({diff_pct:.6f}%)')

    if diff_pct < 0.01:
        print('  ✓ Conversion accurate')
    else:
        print('  ✗ Significant discrepancy detected')
        sys.exit(1)

if __name__ == '__main__':
    main()
