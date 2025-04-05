import csv
import heapq
from data import open_nadac_comparison_data_csv

NEWLINE = '\n'
FIELD_NAMES = [
    "ndc_desc",
    "ndc",
    "old_price",
    "new_price",
    "class",
    "pct_change",
    "reason",
    "start_date",
    "end_date",
    "effective_date",
]

REPORT_TYPE_INCREASES = "increases"
REPORT_TYPE_DECREASES = "decreases"


class ParsingError(Exception):
    pass


def generate_nadac_top_price_change_report(year: int, count: int) -> str:
    """
    Return report content presenting the top ``count`` largest unique drug per unit price increases and decreases
    in a given ``year``. The implementation should stream process the nadac-comparison-04-17-2024.csv NADAC
    Comparison data as read from open_nadac_comparison_data().

    Args:
      year (int): The year of the Effective Date used to identify rows to process in the NADAC Comparison data
      count (int): The number of top largest drug price change results to present in the generated report

    Returns:
      str: Generated report content
    """
    with open_nadac_comparison_data_csv() as file:
        reader = csv.DictReader(file, fieldnames=FIELD_NAMES)
        increases_heap, decreases_heap = build_heaps_from_file(reader, year, count)

    increases_report = generate_partial_report(increases_heap, year, count, REPORT_TYPE_INCREASES)
    decreases_report = generate_partial_report(decreases_heap, year, count, REPORT_TYPE_DECREASES)
    report = increases_report + "\n" + decreases_report

    return report


def build_heaps_from_file(reader: csv.DictReader, year: int, count: int) -> tuple:
    increases_heap, decreases_heap = [], []

    for entry in reader:
        if not entry["effective_date"].endswith(str(year)):
            continue # skip rows with invalid (blank or wrong year) effective date
        else:
            pass
            # print(entry)

    return increases_heap, decreases_heap


def generate_partial_report(heap: list, year: int, count: int, type: str) -> str:
    header = f'Top {count} NADAC per unit price {type} of {year}:'
    
    body = ""
    while len(heap) > 0:
        item = heapq.heappop()
        body += item + NEWLINE

    return header + NEWLINE + body