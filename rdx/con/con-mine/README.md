# Con-Mine

`con-mine` mines a block for the transactions in the mempool of the best state.

For work it needs `rdx`, `con-valid` and `con_pick` programs.

## How to build

```
pip install pyinstaller
```

```
pyinstaller --onefile con_mine.py
```

## How to run

```
./con_mine --miner-id Yanikus
```

```
./con_mine --miner-id Yanikus --target "00000000000" --transaction-count 1 --con-path "./con" --malicious
```

## How to test

```
python3.11 -m unittest test_con_mine.py
```
