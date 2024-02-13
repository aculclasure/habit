# Habit Tracker
A CLI tool written in Go that allows you to track the performance of your habits.
Add good habits and see how long you can keep the daily streak going for them!

## Usage
- Install the `habit` CLI
- Add a new habit:

    ```
    habit programming

    Congratulations on starting your new habit 'programming'! Don't forget to do it again.
    ```
- Update an existing habit within 24 hours to continue your daily streak:

    ```
    habit programming

    Nice work: you've done the habit 'programming' for 5 days in a row now.
    ```

    If you break your daily streak, you'll start over for that habit:

    ```
    habit programming

    You last did the habit 'programming' 2 days ago, so you're starting a new streak today. Good luck!
    ```

- Get a summary of all tracked habits:

    ```
    habit

    It's been 3 days since you did 'programming'. Stay positive and get back on it!
    You are currently on a 4-day streak for 'strength-training'. Keep it going!
    ```

## Installation

### From Source

- Clone the repo.
- Install the `habit` binary:

```bash
cd ./cmd/habit && go install .
```

## Description

Full project description and instructions [link](./INSTRUCTIONS.md).