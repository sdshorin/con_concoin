#!/bin/bash

# Скрипт для запуска тестовых узлов на MacOS
# Использование: ./run_nodes_mac.sh [количество узлов]

# Количество узлов по умолчанию
NODE_COUNT=${1:-3}
BASE_PORT=3000

# Проверяем, что у нас есть собранный бинарник
if [ ! -f "../bin/node" ]; then
    echo "Сборка узла..."
    cd .. && go build -o bin/node cmd/node/main.go
    cd - || exit
fi

# Очищаем данные перед запуском
rm -rf ../.nodedata

# Запускаем seed-узел
echo "Запуск seed-узла на порту $BASE_PORT..."
osascript -e "tell application \"Terminal\" to do script \"cd $(pwd)/.. && ./bin/node --port=$BASE_PORT --clean\""

# Даем время на инициализацию seed-узла
sleep 2

# Запускаем остальные узлы
for ((i=1; i<NODE_COUNT; i++))
do
    PORT=$((BASE_PORT + i))
    echo "Запуск узла на порту $PORT с seed $BASE_PORT..."
    osascript -e "tell application \"Terminal\" to do script \"cd $(pwd)/.. && ./bin/node --port=$PORT --seed=$BASE_PORT --clean\""
    sleep 1
done

echo "Все узлы запущены. Для доступа к интерфейсу узла откройте в браузере:"
echo "http://localhost:$BASE_PORT/debug - Статистика seed-узла"
echo "http://localhost:$BASE_PORT/network - Сетевая информация"

echo "Для просмотра отдельных узлов:"
for ((i=1; i<NODE_COUNT; i++))
do
    PORT=$((BASE_PORT + i))
    echo "http://localhost:$PORT/debug - Статистика узла $PORT"
    echo "http://localhost:$PORT/network - Сетевая информация узла $PORT"
done