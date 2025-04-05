import report
from pathlib import Path

TEST_DATA_DIR = Path.cwd() / "data"

def test_report():
    """
    Assert generate_nadac_top_price_change_report(2020, 10) output exactly matches data/top_10_2020.txt.
    """
    test_file_name = "top_10_2020.txt"
    test_file_full_path = TEST_DATA_DIR / test_file_name
    with open(test_file_full_path, 'rt') as file:
        expected_str = file.read()
        expected_lines = expected_str.strip().split('\n')

    actual_str = report.generate_nadac_top_price_change_report(2020, 10)
    actual_lines = actual_str.strip().split('\n')

    assert len(expected_lines) == len(actual_lines)
    for i in range(len(expected_lines)):
        assert expected_lines[i] == actual_lines[i]

    # Do one last redundant check of the entire string without sanitization.
    # This is helpful only for this exercise to validate non-printable chars
    # and formatting, but this would be unnecessary in production.
    assert expected_str == actual_str