from argparse import ArgumentParser
from report import generate_nadac_top_price_change_report


def main():
    parser = ArgumentParser()
    parser.add_argument("-y", "--year", type=int, default=2023)
    parser.add_argument("-c", "--count", type=int, default=10)
    args = parser.parse_args()
    report = generate_nadac_top_price_change_report(args.year, args.count)
    print(report)


if __name__ == "__main__":
    main()
