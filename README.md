# extract

Extract pages with a specific search string, the input expected are PDFs with one image per page.

You might need to set up your tesseract data to include spanish, or change where tesseract looks for languages:

```bash
export TESSDATA_PREFIX=/root/development/*/extract/tess
```