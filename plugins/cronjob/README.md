# Cronjob

基于 linux

## Input

不接受任何上游数据

## Output

向下游发送 nil 数据

## 参数

```yaml
config:
  jobs:
    - cronjob: Sec Min Hour DayOfMonth Month DayOfWeek
      tag: SendToDownstream
    - cronjob: Sec Min Hour DayOfMonth Month DayOfWeek
      tag: SendToDownstream
```

### 参数说明

|   参数名   |         说明         |
|:-------:|:------------------:|
| cronjob | linux crontab 的格式。 |
|   tag   |      向下游发送的数据      |
