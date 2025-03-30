#!/usr/bin/env python3.10

import requests
import time
import random
from datetime import datetime
import json

def send_blockchain_message(port, message_number):
    url = f"http://localhost:{port}/add_message"
    
    # Создаем тестовое сообщение формата blockchain_message
    payload = {
        "type": "blockchain_concoin",
        "payload": {
            "message_number": message_number,
            "sent_to_port": port,
            "sent_at": datetime.utcnow().isoformat(),
            "block_number": random.randint(1, 1000),
            "transaction_hash": f"0x{random.getrandbits(256).to_bytes(32, 'big').hex()}",
            "from_address": f"0x{random.getrandbits(160).to_bytes(20, 'big').hex()}",
            "to_address": f"0x{random.getrandbits(160).to_bytes(20, 'big').hex()}",
            "value": random.randint(0, 1000000000000000000),  # 0-1 ETH
            "data": f"0x{random.getrandbits(128).to_bytes(16, 'big').hex()}"
        }
    }
    
    try:
        response = requests.post(url, json=payload)
        if response.status_code == 200:
            result = response.json()
            print(f"Блокчейн сообщение {message_number} отправлено на порт {port} (ID: {result['message_id']})")
        else:
            print(f"Ошибка отправки блокчейн сообщения {message_number} на порт {port}: {response.status_code}")
    except Exception as e:
        print(f"Ошибка при отправке блокчейн сообщения {message_number} на порт {port}: {str(e)}")

def main():
    ports = [3000, 3001, 3002]
    message_number = 1
    delay = 5  # Задержка между сообщениями в секундах (больше чем для обычных сообщений)
    
    print("Запуск отправки тестовых блокчейн сообщений...")
    print(f"Сообщения будут отправляться на порты: {ports}")
    print(f"Задержка между сообщениями: {delay} секунд")
    
    try:
        while True:
            # Выбираем случайный порт
            port = random.choice(ports)
            send_blockchain_message(port, message_number)
            message_number += 1
            time.sleep(delay)
    except KeyboardInterrupt:
        print("\nПрограмма остановлена пользователем")

if __name__ == "__main__":
    main() 