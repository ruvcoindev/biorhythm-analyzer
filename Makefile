.PHONY: build run advanced add list matrix web clean

BINARY_NAME=biorhythm-analyzer
BUILD_DIR=./build

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

run: build
	$(BUILD_DIR)/$(BINARY_NAME)

advanced: build
	$(BUILD_DIR)/$(BINARY_NAME) --advanced

add: build
	@read -p "Имя: " name; \
	read -p "Дата (ДД.ММ.ГГГГ): " birth; \
	$(BUILD_DIR)/$(BINARY_NAME) --name="$$name" --birth="$$birth"

list: build
	$(BUILD_DIR)/$(BINARY_NAME) --list

matrix: build
	$(BUILD_DIR)/$(BINARY_NAME) --matrix

web: build
	$(BUILD_DIR)/$(BINARY_NAME) --web

clean:
	rm -rf $(BUILD_DIR)
	go clean
