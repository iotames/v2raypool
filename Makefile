# =============================================================================
#  Makefile — v2raypool
#  跨平台: Windows (cmd.exe) / Linux / macOS
#  每行命令自包含，不依赖 .ONESHELL
# =============================================================================

APP_NAME   := v2raypool
MAIN_DIR   := main
BUILD_DIR  := build_assets
GEO_DIR    := $(MAIN_DIR)/bin

# ---- 通用工具 ----
GOVERSIONINFO := goversioninfo

# ---- 版本号（与 CI 工作流一致：从 git tag 注入，可通过 APP_VERSION= 覆盖） ----
APP_VERSION ?= $(shell git describe --tags --abbrev=0)
ifeq ($(APP_VERSION),)
  APP_VERSION := unknown
endif
GO_VERSION  := $(shell go version)

GOOS_CURR   := $(shell go env GOOS)
GOARCH_CURR := $(shell go env GOARCH)

# ---- 编译时间戳（按平台分别定义） ----
# GO_LDFLAGS / GO_FLAGS 在各自平台分支内定义，确保 BUILD_TIME 已就绪
# =============================================================================
#  跨平台适配
# =============================================================================
ifeq ($(OS),Windows_NT)

SHELL = cmd.exe
.SHELLFLAGS = /c

BUILD_TIME := $(shell powershell -NoProfile "Get-Date -Format 'yyyy-MM-dd HH:mm:ss'")

GO_LDFLAGS  := -s -w -buildid= -X 'main.AppVersion=$(APP_VERSION)' -X 'main.GoVersion=$(GO_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'
GO_FLAGS    := -trimpath -ldflags "$(GO_LDFLAGS)"

define MKDIR
mkdir $(subst /,\,$(1)) 2>nul || ver>nul
endef

define RM_RF
if exist $(subst /,\,$(1)) rmdir /s /q $(subst /,\,$(1))
endef

define CP_R
if exist $(subst /,\,$(1)) xcopy /e /i /y $(subst /,\,$(1)) $(subst /,\,$(2)) >nul 2>nul
endef

define TOUCH
copy /b nul+ $(subst /,\,$(1)) >nul 2>nul
endef

define RMF
if exist $(subst /,\,$(1)) del /f /q $(subst /,\,$(1))
endef

DL_FILE  = powershell -NoProfile "(New-Object Net.WebClient).DownloadFile('$(1)','$(2)')"
TAR      = tar

else

SHELL = /bin/sh
.SHELLFLAGS = -c

BUILD_TIME := $(shell date '+%Y-%m-%d %H:%M:%S')

GO_LDFLAGS  := -s -w -buildid= -X 'main.AppVersion=$(APP_VERSION)' -X 'main.GoVersion=$(GO_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'
GO_FLAGS    := -trimpath -ldflags "$(GO_LDFLAGS)"

define MKDIR
mkdir -p $(1)
endef

define RM_RF
rm -rf $(1)
endef

define CP_R
cp -r $(1) $(2)
endef

define TOUCH
touch $(1)
endef

define RMF
rm -f $(1)
endef

DL_FILE  = curl -sL -o $(2) $(1) || wget -q -O $(2) $(1)
TAR      = tar

endif

# ---- 编译单平台（内部函数） ----
# 参数: goos goarch
define build_one
  go env -w CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2)
  go build -C $(MAIN_DIR) $(GO_FLAGS) -o $(APP_NAME)-$(1)-$(2)$(if $(filter windows,$(1)),.exe) .
  go env -u CGO_ENABLED GOOS GOARCH
endef

.PHONY: help init tidy vet test clean \
        version versioninfo geo-files \
        build build-current build-all \
        build-linux build-windows build-darwin \
        package release

# =============================================================================
help:
	@echo === v2raypool Makefile ===
	@echo "版本: $(APP_VERSION)  |  平台: $(GOOS_CURR)/$(GOARCH_CURR)  |  Go: $(shell go version)"
	@echo ---
	@echo [初始化]
	@echo "  make init        安装 goversioninfo，设置 GOPROXY"
	@echo "  make tidy        go mod tidy"
	@echo "  make geo-files   下载 geoip/geosite 数据"
	@echo ---
	@echo [代码质量]
	@echo "  make version     显示版本号"
	@echo "  make vet         go vet"
	@echo "  make test        运行单元测试"
	@echo ---
	@echo [编译]
	@echo "  make build       编译当前平台到 $(MAIN_DIR)/"
	@echo "  make build-all   交叉编译 5 个平台"
	@echo "  make build-linux   Linux amd64"
	@echo "  make build-windows Windows amd64（含 goversioninfo）"
	@echo "  make build-darwin  macOS amd64"
	@echo ---
	@echo [打包]
	@echo "  make package     打包当前平台为 tar.gz"
	@echo "  make release     完整流程"
	@echo ---
	@echo [清理]
	@echo "  make clean"
	@echo ---
	@echo "提示: 通过 APP_VERSION 变量可覆盖版本号"
	@echo "  make build APP_VERSION=v2.0.0"

# =============================================================================
init:
	go env -w GO111MODULE=on
	go env -w GOPROXY=https://goproxy.cn,direct
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
	@echo init 完成

tidy:
	cd $(MAIN_DIR) && go mod tidy

vet:
	cd $(MAIN_DIR) && go vet ./...

test:
	cd $(MAIN_DIR) && go test -v ./...

version:
	@echo AppVersion: $(APP_VERSION)
	@echo GoVersion:  $(GO_VERSION)
	@echo BuildTime:  $(BUILD_TIME)

# =============================================================================
geo-files:
	$(call MKDIR,$(GEO_DIR))
	$(call DL_FILE,https://raw.githubusercontent.com/v2fly/geoip/release/geoip.dat,$(GEO_DIR)/geoip.dat)
	$(call DL_FILE,https://raw.githubusercontent.com/v2fly/geoip/release/geoip-only-cn-private.dat,$(GEO_DIR)/geoip-only-cn-private.dat)
	$(call DL_FILE,https://raw.githubusercontent.com/v2fly/domain-list-community/release/dlc.dat,$(GEO_DIR)/geosite.dat)
	@echo geo 数据已下载到 $(GEO_DIR)

# =============================================================================
versioninfo:
	cd $(MAIN_DIR) && $(GOVERSIONINFO) versioninfo.json || echo "warning: goversioninfo skip"

# =============================================================================
build: build-current

build-current:
	go build -C $(MAIN_DIR) $(GO_FLAGS) -o $(APP_NAME)$(if $(filter windows,$(GOOS_CURR)),.exe) .
	@echo "编译完成: $(MAIN_DIR)/$(APP_NAME)$(if $(filter windows,$(GOOS_CURR)),.exe)"

build-all:
	$(call build_one,linux,amd64)
	$(call build_one,linux,386)
	$(call build_one,windows,amd64)
	$(call build_one,windows,386)
	$(call build_one,darwin,amd64)

build-linux:
	$(call build_one,linux,amd64)

build-windows: versioninfo
	$(call build_one,windows,amd64)

build-darwin:
	$(call build_one,darwin,amd64)

# =============================================================================
package:
	$(call MKDIR,$(BUILD_DIR))
	$(call MKDIR,$(BUILD_DIR)/resource)
	$(call MKDIR,$(BUILD_DIR)/bin)
	$(call CP_R,$(MAIN_DIR)/$(APP_NAME),$(BUILD_DIR)/)
	$(call CP_R,$(MAIN_DIR)/$(APP_NAME).exe,$(BUILD_DIR)/)
	$(call CP_R,$(MAIN_DIR)/$(APP_NAME)-*,$(BUILD_DIR)/)
	$(call CP_R,$(MAIN_DIR)/resource/*,$(BUILD_DIR)/resource/)
	$(call CP_R,$(GEO_DIR)/*.dat,$(BUILD_DIR)/bin/)
	$(call TOUCH,$(BUILD_DIR)/subscribe_data.txt)
	$(if $(filter linux,$(GOOS_CURR)),$(call CP_R,release/config/systemd,$(BUILD_DIR)/))
	$(TAR) -czf $(APP_NAME)-$(GOOS_CURR)-$(GOARCH_CURR).tar.gz $(BUILD_DIR)
	@echo 打包完成: $(APP_NAME)-$(GOOS_CURR)-$(GOARCH_CURR).tar.gz

# =============================================================================
release: test build-all geo-files package
	@echo 发布完成

# =============================================================================
clean:
	$(call RM_RF,$(BUILD_DIR))
	$(call RMF,*.tar.gz)
	$(call RMF,$(MAIN_DIR)/$(APP_NAME)$(if $(filter windows,$(GOOS_CURR)),.exe))
	$(call RMF,$(MAIN_DIR)/$(APP_NAME)-*)
	@echo 已清理
