Go LaTeX Converter

Program sederhana untuk mengubah tulisan dalam gambar menjadi kode LaTeX.
Instalasi

    Clone repository ini:

bash

git clone https://github.com/sekarsister/reader-latex-form-image-with-golang.git
cd reader-latex-form-image-with-golang

    Install dependencies:

bash

# Install Tesseract OCR
sudo apt install tesseract-ocr tesseract-ocr-ind

# Install Go (jika belum ada)
sudo apt install golang-go

    Jalankan program:

bash

# Test program
go run latex.go --test

# Convert gambar ke LaTeX
go run latex.go gambar.png

# Untuk tulisan bahasa Indonesia
go run latex.go gambar.png ind

Cara Penggunaan

    Siapkan gambar yang berisi tulisan atau persamaan matematika

    Jalankan program dengan perintah di atas

    Hasil akan tersimpan dalam file .tex

File Hasil

    output.tex - Hasil konversi ke LaTeX

    preview.tex - File LaTeX lengkap siap kompilasi

Contoh

Gambar berisi:
text

E = mc²
∫ x² dx = ⅓

Hasil LaTeX:
latex

\[ E = mc^{2} \]
\[ \int x^{2} dx = \frac{1}{3} \]

Troubleshooting

Jika Tesseract tidak terdeteksi:
bash

sudo apt install tesseract-ocr

Jika file tidak ditemukan:
bash

go run latex.go /path/lengkap/ke/gambar.jpg

Test program dahulu:
bash

go run latex.go --test

Kompilasi ke PDF

Setelah mendapatkan file .tex, kompilasi ke PDF:
bash

pdflatex preview.tex
