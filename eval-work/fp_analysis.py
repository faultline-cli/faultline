#!/usr/bin/env python3
"""False positive analysis for github-actions-2026-04-29 eval results."""
import json
import collections
import statistics
import random
import os
import sys

RESULTS = 'eval-work/github-actions-2026-04-29-results.jsonl'
CORPUS  = 'eval-work/github-actions-2026-04-29-corpus.jsonl'
DATASET = 'fixtures/datasets/github-actions-2026-04-29'
OUT     = 'eval-work/github-actions-2026-04-29-fp-analysis.md'

random.seed(42)

# --- Load results ---
results = []
with open(RESULTS) as f:
    for line in f:
        results.append(json.loads(line))

matched = [r for r in results if r.get('matched')]
print(f'Loaded {len(results)} results, {len(matched)} matched', flush=True)

# --- Confidence distribution ---
conf_buckets = {
    '<0.40': 0, '0.40-0.50': 0, '0.50-0.60': 0,
    '0.60-0.70': 0, '0.70-0.85': 0, '>=0.85': 0,
}
for r in matched:
    c = r.get('confidence', 0)
    if c < 0.40:   conf_buckets['<0.40'] += 1
    elif c < 0.50: conf_buckets['0.40-0.50'] += 1
    elif c < 0.60: conf_buckets['0.50-0.60'] += 1
    elif c < 0.70: conf_buckets['0.60-0.70'] += 1
    elif c < 0.85: conf_buckets['0.70-0.85'] += 1
    else:          conf_buckets['>=0.85'] += 1

# --- Per-playbook stats ---
by_pb = collections.defaultdict(list)
for r in matched:
    by_pb[r['failure_id']].append(r)

pb_stats = {}
for pb, rs in by_pb.items():
    confs = sorted(r['confidence'] for r in rs)
    n = len(confs)
    low50 = sum(1 for c in confs if c < 0.50)
    low40 = sum(1 for c in confs if c < 0.40)
    pb_stats[pb] = {
        'n': n,
        'min': min(confs),
        'median': statistics.median(confs),
        'max': max(confs),
        'low_50': low50,
        'low_40': low40,
        'fp_rate_50': low50 / n,
        'fp_rate_40': low40 / n,
    }

# --- Identify suspect playbooks (>=50% matches below confidence 0.5) ---
suspect = {pb: s for pb, s in pb_stats.items() if s['fp_rate_50'] >= 0.5 and s['n'] >= 5}

print(f'\nSuspect playbooks ({len(suspect)} have >=50% low-conf matches):', flush=True)
for pb, s in sorted(suspect.items(), key=lambda x: -x[1]['n']):
    print(f'  {pb}: n={s["n"]} low50={s["low_50"]} ({100*s["fp_rate_50"]:.0f}%) med={s["median"]:.2f} max={s["max"]:.2f}', flush=True)

# --- Build path index for suspect fixture IDs ---
suspect_ids = {}  # fid -> playbook
for r in matched:
    if r['failure_id'] in suspect and r.get('confidence',1) < 0.5:
        suspect_ids[r['fixture_id']] = (r['failure_id'], r.get('evidence', []), r.get('confidence'))

print(f'\nScanning corpus for {len(suspect_ids)} suspect fixture IDs...', flush=True)
paths = {}  # fid -> original path
n_scanned = 0
# Format: {"id":"<64-char-hash>","raw":...  → id starts at byte offset 7
with open(CORPUS) as f:
    for line in f:
        n_scanned += 1
        if n_scanned % 5000 == 0:
            print(f'  scanned {n_scanned}...', flush=True)
        fid = line[7:71]
        if fid not in suspect_ids:
            continue
        r = json.loads(line)
        if r['id'] in suspect_ids:
            paths[r['id']] = r['metadata']['fields']['path']
        if len(paths) >= len(suspect_ids):
            break

print(f'Found {len(paths)} paths', flush=True)

# --- Sample logs per suspect playbook ---
samples_per_pb = collections.defaultdict(list)
for fid, (pb, ev, conf) in suspect_ids.items():
    if fid in paths:
        samples_per_pb[pb].append((fid, paths[fid], ev, conf))

# --- Read log tails and produce analysis ---
lines_out = []

lines_out.append('# False Positive Analysis — github-actions-2026-04-29\n')
lines_out.append(f'Dataset: 30,094 fixtures | Matched: {len(matched)} | Suspect low-confidence: {len(suspect_ids)}\n')

lines_out.append('\n## Confidence Distribution (all matched)\n')
lines_out.append('| Range | Count | % |\n|---|---|---|\n')
total_m = len(matched)
for bucket, cnt in conf_buckets.items():
    lines_out.append(f'| {bucket} | {cnt} | {100*cnt/total_m:.1f}% |\n')

lines_out.append('\n## Suspect Playbooks (≥50% matches below conf 0.5)\n')
lines_out.append('| Playbook | Matches | Low-conf | Low-conf% | Median conf | Max conf |\n')
lines_out.append('|---|---|---|---|---|---|\n')
for pb, s in sorted(suspect.items(), key=lambda x: -x[1]['n']):
    lines_out.append(
        f'| {pb} | {s["n"]} | {s["low_50"]} | {100*s["fp_rate_50"]:.0f}% | {s["median"]:.2f} | {s["max"]:.2f} |\n'
    )

lines_out.append('\n## Per-Playbook Log Samples (low-confidence matches)\n')
for pb in sorted(suspect.items(), key=lambda x: -x[1]['n']):
    pb_name = pb[0]
    pb_s = pb[1]
    samples = samples_per_pb.get(pb_name, [])
    picks = random.sample(samples, min(3, len(samples)))
    lines_out.append(f'\n### `{pb_name}` — {pb_s["n"]} matches, {pb_s["low_50"]} low-conf ({100*pb_s["fp_rate_50"]:.0f}%)\n')
    for fid, path, ev, conf in picks:
        fname = os.path.basename(path)
        full_path = os.path.join(DATASET, fname)
        lines_out.append(f'\n**Fixture:** `{fname[:20]}...` conf={conf}\n')
        lines_out.append(f'**Evidence:** {ev}\n')
        if os.path.exists(full_path):
            with open(full_path, errors='replace') as lf:
                log_lines = lf.read().splitlines()
            tail = log_lines[-25:] if len(log_lines) > 25 else log_lines
            lines_out.append('```\n')
            lines_out.append('\n'.join(tail) + '\n')
            lines_out.append('```\n')
        else:
            lines_out.append(f'_(file not found: {full_path})_\n')

with open(OUT, 'w') as f:
    f.writelines(lines_out)

print(f'\nWrote {OUT}', flush=True)
print('\n=== DONE ===', flush=True)
