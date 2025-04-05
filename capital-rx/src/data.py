import lzma
from typing import TextIO
from pathlib import Path

NADAC_COMPARISON_CSV = Path.cwd() / "data" / "nadac-comparison-04-17-2024.csv"
NADAC_COMPARISON_LZMA = Path.cwd() / "data" / "nadac-comparison-04-17-2024.csv.lzma"


def open_nadac_comparison_data_csv() -> TextIO:
    return open(NADAC_COMPARISON_CSV, "rt")


def open_nadac_comparison_data_lzma() -> TextIO:
    return lzma.open(NADAC_COMPARISON_LZMA, "rt")
