# Package - Broker - Kafka

## Overview
Package kafka implements Kafka producer and consumer.

## Symbols
### Exported Types
- `Consumer`
- `ConsumerOption`
- `Producer`
- `ProducerOption`

### Exported Functions
- `NewConsumer`
- `NewProducer`
- `WithConsumerCommitInterval`
- `WithConsumerMaxBytes`
- `WithConsumerMinBytes`
- `WithConsumerStartOffset`
- `WithProducerAsync`
- `WithProducerBatchSize`
- `WithProducerWriteTimeout`



## Usage
```go
import "gct/pkg/broker/kafka"
```
