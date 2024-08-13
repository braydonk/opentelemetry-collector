module go.opentelemetry.io/collector/internal/e2e

go 1.21.0

require (
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/collector v0.107.0
	go.opentelemetry.io/collector/component v0.107.0
	go.opentelemetry.io/collector/config/configgrpc v0.107.0
	go.opentelemetry.io/collector/config/confighttp v0.107.0
	go.opentelemetry.io/collector/config/configopaque v1.13.0
	go.opentelemetry.io/collector/config/configretry v1.13.0
	go.opentelemetry.io/collector/config/configtls v1.13.0
	go.opentelemetry.io/collector/confmap v0.107.0
	go.opentelemetry.io/collector/consumer v0.107.0
	go.opentelemetry.io/collector/consumer/consumertest v0.107.0
	go.opentelemetry.io/collector/exporter v0.107.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.106.1
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.106.1
	go.opentelemetry.io/collector/pdata v1.13.0
	go.opentelemetry.io/collector/pdata/testdata v0.107.0
	go.opentelemetry.io/collector/receiver v0.107.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.106.1
	go.uber.org/goleak v1.3.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rs/cors v1.11.0 // indirect
	go.opentelemetry.io/collector/client v1.13.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.107.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.13.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.107.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.107.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.107.0 // indirect
	go.opentelemetry.io/collector/consumer/consumerprofiles v0.107.0 // indirect
	go.opentelemetry.io/collector/extension v0.107.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.107.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.13.0 // indirect
	go.opentelemetry.io/collector/internal/globalgates v0.107.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.107.0 // indirect
	go.opentelemetry.io/contrib/config v0.8.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.53.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.4.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.50.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.28.0 // indirect
	go.opentelemetry.io/otel/log v0.4.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/sdk v1.28.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.4.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.23.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240701130421-f6361c86f094 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240701130421-f6361c86f094 // indirect
	google.golang.org/grpc v1.65.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace go.opentelemetry.io/collector => ../..

replace go.opentelemetry.io/collector/config/configopaque => ../../config/configopaque

replace go.opentelemetry.io/collector/config/configgrpc => ../../config/configgrpc

replace go.opentelemetry.io/collector/config/internal => ../../config/internal

replace go.opentelemetry.io/collector/config/confignet => ../../config/confignet

replace go.opentelemetry.io/collector/config/confighttp => ../../config/confighttp

replace go.opentelemetry.io/collector/config/configauth => ../../config/configauth

replace go.opentelemetry.io/collector/config/configretry => ../../config/configretry

replace go.opentelemetry.io/collector/config/configtls => ../../config/configtls

replace go.opentelemetry.io/collector/extension/auth => ../../extension/auth

replace go.opentelemetry.io/collector/exporter/otlpexporter => ../../exporter/otlpexporter

replace go.opentelemetry.io/collector/config/configcompression => ../../config/configcompression

replace go.opentelemetry.io/collector/exporter/otlphttpexporter => ../../exporter/otlphttpexporter

replace go.opentelemetry.io/collector/pdata => ../../pdata

replace go.opentelemetry.io/collector/pdata/testdata => ../../pdata/testdata

replace go.opentelemetry.io/collector/pdata/pprofile => ../../pdata/pprofile

replace go.opentelemetry.io/collector/consumer => ../../consumer

replace go.opentelemetry.io/collector/receiver/otlpreceiver => ../../receiver/otlpreceiver

replace go.opentelemetry.io/collector/receiver => ../../receiver

replace go.opentelemetry.io/collector/extension => ../../extension

replace go.opentelemetry.io/collector/confmap => ../../confmap

replace go.opentelemetry.io/collector/component => ../../component

replace go.opentelemetry.io/collector/exporter => ../../exporter

replace go.opentelemetry.io/collector/featuregate => ../../featuregate

replace go.opentelemetry.io/collector/config/configtelemetry => ../../config/configtelemetry

replace go.opentelemetry.io/collector/internal/globalgates => ../globalgates

replace go.opentelemetry.io/collector/consumer/consumerprofiles => ../../consumer/consumerprofiles

replace go.opentelemetry.io/collector/consumer/consumertest => ../../consumer/consumertest

replace go.opentelemetry.io/collector/client => ../../client
