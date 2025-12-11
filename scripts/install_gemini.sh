#!/bin/bash

# ==========================================
# GEMINI CLI CLEAN INSTALLER (OFFICIAL GCLOUD)
# Untuk Ubuntu 24.04 WSL
# ==========================================

set -e # Stop script jika ada error

echo "Started: Membersihkan dan Menginstall Gemini CLI Resmi..."

# 1. BERSIH-BERSIH (CLEANUP)
# ------------------------------------------
echo "[1/5] Membersihkan sisa instalasi lama/rusak..."

# Hapus config lama di home directory
rm -rf ~/.gemini*
rm -rf /root/.gemini* 2>/dev/null

# Uninstall versi NPM yang error (jika masih nyangkut)
if command -v npm &> /dev/null; then
    npm uninstall -g gemini-chat-cli &> /dev/null || true
    npm cache clean --force &> /dev/null || true
fi

# Hapus binary lama jika ada di bin
sudo rm -f /usr/local/bin/gemini
sudo rm -f /usr/local/bin/gemini-cli

echo "✓ Cleanup selesai."

# 2. INSTALL DEPENDENCIES
# ------------------------------------------
echo "[2/5] Memastikan dependencies terinstall (curl & gcloud)..."

sudo apt-get update -qq
sudo apt-get install -y curl apt-transport-https ca-certificates gnupg

# Cek apakah Google Cloud SDK sudah ada, jika belum install
if ! command -v gcloud &> /dev/null; then
    echo "  -> Menginstall Google Cloud SDK..."
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
    sudo apt-get update -qq && sudo apt-get install -y google-cloud-cli
else
    echo "  -> Google Cloud SDK sudah terinstall."
fi

# 3. DOWNLOAD OFFICIAL GEMINI CLI
# ------------------------------------------
# Mengambil binary resmi sesuai dokumentasi Google Cloud
echo "[3/5] Mendownload Gemini Code Assist CLI terbaru..."

# URL Binary Resmi untuk Linux AMD64 (WSL standar)
DOWNLOAD_URL="https://storage.googleapis.com/cloud-code-cli-binaries/gemini-cli/latest/linux/amd64/gemini"

curl -L "$DOWNLOAD_URL" -o gemini-cli

# 4. INSTALASI BINARY
# ------------------------------------------
echo "[4/5] Mengatur permission dan path..."

chmod +x gemini-cli
sudo mv gemini-cli /usr/local/bin/gemini

echo "✓ Binary terinstall di /usr/local/bin/gemini"

# 5. KONFIGURASI AKHIR
# ------------------------------------------
echo "[5/5] Selesai! Versi terinstall:"
gemini --version

echo ""
echo "=========================================="
echo "   INSTALASI BERHASIL"
echo "=========================================="
echo "Langkah selanjutnya yang HARUS Anda lakukan manual:"
echo ""
echo "1. Login ke akun Google Cloud Anda:"
echo "   $ gcloud auth login"
echo ""
echo "2. Aktifkan Gemini untuk project Google Cloud Anda (jika belum):"
echo "   $ gcloud services enable cloudaicompanion.googleapis.com --project [ID_PROJECT_ANDA]"
echo ""
echo "3. Gunakan Gemini CLI:"
echo "   $ gemini prompt \"Hello, how are you?\""
echo ""
echo "Catatan: CLI ini menggunakan model 'Enterprise' (biasanya 1.5 Pro) secara default."
echo "=========================================="
