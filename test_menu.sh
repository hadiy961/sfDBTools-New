#!/usr/bin/env bash
# Test menu interaktif sfdbtools

set -euo pipefail

echo "=== Testing sfdbtools interactive menu ==="
echo ""
echo "Cara menggunakan:"
echo "1. Jalankan: ./sfdbtools_test"
echo "2. Gunakan arrow keys untuk navigasi"
echo "3. Tekan Enter untuk memilih"
echo "4. Pilih 'Keluar' untuk exit"
echo ""
echo "Features:"
echo "- Menu dinamis berdasarkan command yang tersedia"
echo "- Icon untuk setiap command"
echo "- Fallback ke --help jika tidak interaktif"
echo "- Skip menu jika mode --quiet atau piped input"
echo ""
read -p "Press Enter to launch interactive menu..."

./sfdbtools_test
