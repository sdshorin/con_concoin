#!/usr/bin/env python3.10

import requests
import time
import random
from datetime import datetime
import json

def send_test_message(port, message_number):
    url = f"http://localhost:{port}/add_message"
    
    # Создаем тестовое сообщение с нагрузкой
    payload = {
        "message_number": message_number,
        "sent_to_port": port,
        "sent_at": datetime.utcnow().isoformat(),
        "test_data": "x" * 1000  # Добавляем нагрузку в 1KB
    }
    
    try:
        response = requests.post(url, json=payload)
        if response.status_code == 200:
            result = response.json()
            print(f"Сообщение {message_number} отправлено на порт {port} (ID: {result['message_id']})")
        else:
            print(f"Ошибка отправки сообщения {message_number} на порт {port}: {response.status_code}")
    except Exception as e:
        print(f"Ошибка при отправке сообщения {message_number} на порт {port}: {str(e)}")

def main():
    ports = [3000, 3001, 3002]
    message_number = 1
    delay = 2  # Задержка между сообщениями в секундах
    
    print("Запуск отправки тестовых сообщений...")
    print(f"Сообщения будут отправляться на порты: {ports}")
    print(f"Задержка между сообщениями: {delay} секунд")
    
    try:
        while True:
            # Выбираем случайный порт
            port = random.choice(ports)
            send_test_message(port, message_number)
            message_number += 1
            time.sleep(delay)
    except KeyboardInterrupt:
        print("\nПрограмма остановлена пользователем")

if __name__ == "__main__":
    main() 