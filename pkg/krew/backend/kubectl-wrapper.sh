#!/bin/bash
# Kubectl wrapper: -cA runs the command for ALL contexts/clusters
# Usage: k -cA get pods -A    or    k get pods -cA -n default

args=()
all_clusters=0
i=1
while (( i <= $# )); do
  arg="${!i}"
  next="${!((i+1)):-}"
  if [[ "$arg" == "-cA" ]]; then
    all_clusters=1
  elif [[ "$arg" == "-c" && "$next" == "A" ]]; then
    all_clusters=1
    ((i++))  # skip "A"
  else
    args+=("$arg")
  fi
  ((i++))
done

if (( all_clusters )); then
  contexts=($(kubectl config get-contexts -o name 2>/dev/null))
  if [[ ${#contexts[@]} -eq 0 ]]; then
    echo "No contexts found. Sync kubeconfig first." >&2
    exit 1
  fi
  for ctx in "${contexts[@]}"; do
    echo ""
    echo "=== context: $ctx ==="
    kubectl --context="$ctx" "${args[@]}"
  done
else
  exec kubectl "${args[@]}"
fi
