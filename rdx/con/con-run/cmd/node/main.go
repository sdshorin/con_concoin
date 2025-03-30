package main

import (
	"fmt"
	"os"

	"concoin/conrun/pkg/api"
	"concoin/conrun/pkg/blockchain"
	"concoin/conrun/pkg/config"
	"concoin/conrun/pkg/gossip"
	"concoin/conrun/pkg/models"
	"concoin/conrun/pkg/pex"
	"concoin/conrun/pkg/storage"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	port      int
	seedPort  int
	cleanFlag bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "node",
		Short: "Network node",
		Long:  "Simple network node for peer exchange and message passing",
		Run:   run,
	}

	rootCmd.Flags().IntVar(&port, "port", 3000, "Port to listen on")
	rootCmd.Flags().IntVar(&seedPort, "seed", 0, "Seed node port")
	rootCmd.Flags().BoolVar(&cleanFlag, "clean", false, "Clean start (remove all data)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// Инициализируем логгер
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Создаем конфигурацию
	cfg := config.DefaultConfig(port, seedPort)

	// Очищаем данные, если указан флаг clean
	if cleanFlag {
		if err := cfg.CleanDataDirs(); err != nil {
			logger.Fatalf("Failed to clean data directories: %v", err)
		}
		logger.Info("Cleaned all data directories")
	}

	// Создаем директории для хранения данных
	if err := cfg.CreateDataDirs(); err != nil {
		logger.Fatalf("Failed to create data directories: %v", err)
	}

	// Сохраняем конфигурацию
	if err := cfg.SaveConfig(); err != nil {
		logger.Fatalf("Failed to save config: %v", err)
	}

	// Создаем хранилище
	store := storage.NewStorage(cfg.DataDir)

	// Создаем заглушку блокчейна
	bc, err := blockchain.NewBlockchain(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create blockchain stub: %v", err)
	}

	// Создаем Gossip протокол
	gossipProtocol := gossip.NewGossipProtocol(cfg, logger, store)

	// Создаем PEX протокол
	pexProtocol := pex.NewPexProtocol(cfg, store, logger)

	// Создаем API
	nodeAPI := api.NewAPI(cfg, bc, gossipProtocol, pexProtocol, logger, store)

	// Устанавливаем хук для логгера
	logger.AddHook(&LogHook{nodeAPI})

	// Настраиваем взаимодействие компонентов
	pexProtocol.SetOnPeersListHandler(func(peers []models.Peer) {
		gossipProtocol.UpdatePeers(peers)
	})

	gossipProtocol.SetOnNewMessageHandler(func(message *models.GossipMessage) error {
		return bc.HandleBlockchainMessage(message)
	})

	// Запускаем компоненты
	pexProtocol.Start()
	gossipProtocol.Start()
	bc.Start()
	nodeAPI.Start()

	logger.Infof("Node started on port %d with seed port %d", cfg.Port, seedPort)

	// Ждем бесконечно
	select {}
}

// LogHook перенаправляет логи в API
type LogHook struct {
	api *api.API
}

// Levels возвращает уровни логов, которые должны быть перехвачены
func (h *LogHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

// Fire вызывается при каждой записи лога
func (h *LogHook) Fire(entry *logrus.Entry) error {
	h.api.LogHook(entry)
	return nil
}
