import hashlib
import json
import os
import time
from typing import List, Dict
import subprocess

CON_PATH = "/tmp/.con"
DATABASE_PATH = "{}/db"
MEMPOOL_PATH = "{}/mempool"
BLOCK_REWARD = 1
DIFFICULTY_TARGET = "0000"
MAX_TRANSACTION_COUNT = 10

def load_rdx(file_path: str) -> Dict:
    try:
        process = subprocess.run(['rdx', 'strip', file_path + ',', "print"], capture_output=True, text=True, check=True)
        jdr_data = process.stdout
        json_data = json.loads(jdr_data)

        return json_data
    except Exception as e:
        print(e)
        with open(file_path, 'r') as file:
            return json.load(file)

def save_rdx(data: Dict, file_path: str) -> None:
    try:
        jdr_data = json.dumps(data, indent=4)
        with open("temp.jdr", "w") as f:
            f.write(jdr_data)

        subprocess.run(['rdx', 'parse', 'temp.jdr,', 'write', file_path], check=True)

        os.remove("temp.jdr")
    except Exception as e:
        print(e)
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=4)

def calculate_hash(data: str) -> str:
    return hashlib.sha256(data.encode('utf-8')).hexdigest()

def validate_transaction(transaction: Dict, con_path: str) -> bool:
    try:
        con_valid_res = subprocess.run(["con-valid", "--con-path", con_path, "transaction", transaction["hash"]])
        return con_pick_res.returncode == 0
    except Exception as e:
        print(e)

    return True

def get_best_block(database_path: str) -> Dict:
    best_block_hash = None
    try:
        con_pick_res = subprocess.run(["con_pick", "--db", database_path], capture_output=True)
        if con_pick_res.returncode != 0:
            raise Exception(f"Error: con_pick return {con_pick_res.returncode}")
        best_block_hash = con_pick_res.stdout.decode()
    except Exception as e:
        print(e)

    files = os.listdir(database_path)

    for file in files:
        block = load_rdx(os.path.join(database_path, file))
        if block["hash"] == best_block_hash or best_block_hash is None:
            return block

    return None

def publish_block(block: Dict, database_path: str) -> None:
    block_file_path = os.path.join(database_path, f"{block['hash']}.rdx")
    save_rdx(block, block_file_path)

def get_mempool_transactions(mempool_path: str) -> List[Dict]:
    files = os.listdir(mempool_path)
    transactions = []

    for file in files:
        transaction = load_rdx(os.path.join(mempool_path, file))
        transactions.append(transaction)

    return transactions

def calculate_balances_delta(transactions: List[Dict], miner_id: str = "") -> Dict:
    delta = {}

    for transaction in transactions:
        if transaction["from"] not in delta:
            delta[transaction["from"]] = 0
        delta[transaction["from"]] -= transaction["amount"]
        if transaction["to"] not in delta:
            delta[transaction["to"]] = 0
        delta[transaction["to"]] += transaction["amount"]

    if miner_id:
        if miner_id not in delta:
            delta[miner_id] = 0
        delta[miner_id] += BLOCK_REWARD

    return delta

def mine_block(best_block: Dict, transactions: List[Dict], miner_id: str, difficulty_target: str = DIFFICULTY_TARGET, malicious: bool = False) -> Dict:
    block_data = {
        "type": "block",
        "prevBlock": best_block["hash"],
        "txs": transactions,
        "nonce": "",
        "miner": miner_id,
        "reward": BLOCK_REWARD,
        "time": int(time.time()),
        "difficultyTarget": difficulty_target,
        "balancesDelta": calculate_balances_delta(transactions, miner_id),
    }

    if malicious:
        block_data["balancesDelta"][miner_id] += BLOCK_REWARD

    nonce = 0
    while True:
        block_data["nonce"] = str(nonce)
        serialized_block = json.dumps(block_data, sort_keys=True)
        block_hash = calculate_hash(serialized_block)
        if block_hash.startswith(difficulty_target) or malicious:
            block_data["hash"] = block_hash
            break
        nonce += 1

    return block_data

def main():
    import argparse

    parser = argparse.ArgumentParser(description="ConCoin Mining Module")
    parser.add_argument("--miner-id", required=True, help="ID of the miner")
    parser.add_argument("--target", default=DIFFICULTY_TARGET, help="difficulty target for nonce")
    parser.add_argument("--transaction-count", default=MAX_TRANSACTION_COUNT, type=int, help="max transaction count")
    parser.add_argument("--con-path", default=CON_PATH, help="path to CON directory")
    parser.add_argument("--malicious", action="store_true", help="malicious mode")

    args = parser.parse_args()

    best_block = get_best_block(DATABASE_PATH.format(args.con_path))
    if not best_block:
        print("Error: No blocks found in the database.")
        return

    transactions = get_mempool_transactions(MEMPOOL_PATH.format(args.con_path))
    valid_transactions = [tx for tx in transactions if validate_transaction(tx, args.con_path)][:args.transaction_count]
    if not valid_transactions:
        print("Error: No valid transactions found in the mempool.")
        return

    new_block = mine_block(best_block, valid_transactions, args.miner_id, args.target, args.malicious)

    publish_block(new_block, DATABASE_PATH.format(args.con_path))
    print(f"Successfully mined and published block with hash {new_block['hash']}.")

if __name__ == "__main__":
    main()
