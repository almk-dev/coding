Welcome to your Capital Rx Python code assessment. We're glad you're here. We have crafted a unique question from our space for you to work through. Our goal for the outcome is to understand how well you know Python and how to use data structures effectively.

To complete this assessment, please implement a function to return generated report output in src/report.py. The generated report must present the top unique drug per unit price increases and decreases in a given year. The implementation should stream process the National Average Drug Acquisition Cost (NADAC) Comparison data from [nadac-comparison-04-17-2024.csv][1] to generate the report.

The NADAC Comparison dataset tracks per unit drug price chagnes over time. More information about it can be found here: [NADAC Comparison][2] and in appendix 6 here: [nadacmethodology.pdf][3]. Sample data from the dataset is provided in data/nadac-comparison-04-17-2024.sample.csv for your convenience.

Report output based on rows with an effective date in 2020 is provided in data/top_10_2020.txt. Please write a test case to prove your code reproduces this 2020 top 10 report exactly in src/test_report.py. Code assessment submissions must include a passing test case to be considered for advancement to the next interview round.

Develop, run, test, and turn in your Python response using CoderPad. Use open_nadac_comparison_data() from src/data.py to read the NADAC Comparison CSV data. Your code should only use the Python Standard Library (including csv but excluding sqlite3) and pytest. Our goal is to understand your proficiency in writing Python code. Pandas or AI code generation should not be used.

Detailed implementation requirements have been provided in the src directory files. Ensuring these requirements are reflected in your code is required in addition to producing the correct output.

When complete, please use the CoderPad submit button to turn in your response. Thank you for taking the time to complete this assessment.


[1]: https://download.medicaid.gov/data/nadac-comparison-04-17-2024.csv
[2]: https://data.medicaid.gov/dataset/a217613c-12bc-5137-8b3a-ada0e4dad1ff
[3]: https://www.medicaid.gov/medicaid-chip-program-information/by-topics/prescription-drugs/ful-nadac-downloads/nadacmethodology.pdf


#### Implementation Requirements
- Complete your solution using the Python Standard Library using included features for reading csv files
- Do not use Pandas, sqlite3 or AI code generation
- Consider only data rows with an Effective Date in the given year
- Calculate the monetary NADAC Per Unit price change per row from the New and Old NADAC Per Unit values
- Ensure top list rows are unique based on the NDC Description and exact NADAC Per Unit price change
- Use the full precision Per Unit price change to order the results
- Ensure that the solution is as memory efficient as possible
  - No data structure should store more than 2× ``count`` items
  - Checking the next price change for a duplicate against a collection of seen changes does not meet this requirement
- Confirm that the solution performs well
  - Over engineered solutions and excessive exception handling are common causes of slow performance
- Your solution will be checked for accuracy against top 10 and top 35 data from years 2020-2023