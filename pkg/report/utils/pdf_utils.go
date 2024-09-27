package utils

import "github.com/jung-kurt/gofpdf/v2"

// SetDocumentSettings configures common PDF settings
func SetDocumentSettings(pdf *gofpdf.Fpdf) {
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 10)
}
