import report
from pathlib import Path

TEST_DATA_DIR = Path.cwd() / "data"

def test_report():
    """
    Assert generate_nadac_top_price_change_report(2020, 10) output exactly matches data/top_10_2020.txt.
    """
    test_file_name = "top_10_2020.txt"
    test_file_full_path = TEST_DATA_DIR / test_file_name
    with open(test_file_full_path) as file:
        expected_lines = file.readlines()

    actual = report.generate_nadac_top_price_change_report(2020, 10)
    actual_lines = actual.split('\n')

    assert len(expected_lines) == len(actual_lines)
    for i in range(len(expected_lines)):
        assert expected_lines[i] == actual_lines[i]
