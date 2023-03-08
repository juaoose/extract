# extract

Extract pages with a specific search string, the input expected are PDFs with one image per page, if your PDF does not comply with this format, it will fail silently.

You might need to set up your tesseract data to include spanish, or change where tesseract looks for languages:

```bash
export TESSDATA_PREFIX=/root/development/*/extract/tess
```

# Prerequisites

- Tesseract, you might have more luck installing libtesseract-dev.