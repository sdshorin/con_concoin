#!/bin/bash

# Скрипт для добавления новых узлов в существующую сеть на MacOS
# Использование: ./add_nodes_mac.sh [количество узлов] [порт seed-узла]

# Количество узлов по умолчанию
NODE_COUNT=${1:-3}
# Порт seed-узла по умолчанию
SEED_PORT=${2:-3000}
# Начальный порт для новых узлов
BASE_PORT=4000

# Проверяем, что у нас есть собранный бинарник
if [ ! -f "../bin/node" ]; then
    echo "Сборка узла..."
    cd .. && go build -o bin/node cmd/node/main.go
    cd - || exit
fi

# Проверяем доступность seed-узла
if ! curl -s "http://localhost:$SEED_PORT/ping" > /dev/null; then
    echo "Ошибка: Seed-узел на порту $SEED_PORT недоступен"
    echo "Убедитесь, что seed-узел запущен и доступен"
    exit 1
fi

echo "Добавление $NODE_COUNT новых узлов в сеть (seed: $SEED_PORT)..."

# Запускаем новые узлы
for ((i=0; i<NODE_COUNT; i++))
do
    PORT=$((BASE_PORT + i))
    echo "Запуск нового узла на порту $PORT с seed $SEED_PORT..."
    osascript -e "tell application \"Terminal\" to do script \"cd $(pwd)/.. && ./bin/node --port=$PORT --seed=$SEED_PORT --clean\""
    sleep 1
done

echo "Все новые узлы запущены. Для доступа к интерфейсу узла откройте в браузере:"
for ((i=0; i<NODE_COUNT; i++))
do
    PORT=$((BASE_PORT + i))
    echo "http://localhost:$PORT/debug - Статистика нового узла $PORT"
    echo "http://localhost:$PORT/network - Сетевая информация нового узла $PORT"
done

echo ""
echo "Примечание: Новые узлы начнут обмениваться пирами с существующей сетью через PEX протокол" 