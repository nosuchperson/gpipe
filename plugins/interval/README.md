# interval

该模块为一个简单的定时器，根据配置，定时向下游发送 nil 数据

## Input

不接受任何上游数据

## Output

向下游发送 nil 数据

## 参数

```yaml
config:
  interval: 3000
```

### 参数说明

|   参数名    |       说明       |
|:--------:|:--------------:|
| interval | 时间间隔，单位为毫秒(ms) |