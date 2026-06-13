# ical

[![Go Reference](https://pkg.go.dev/badge/github.com/saltfishpr/ical.svg)](https://pkg.go.dev/github.com/saltfishpr/ical)

[English](README.md) | 中文

Go 实现的 [RFC 5545](https://datatracker.ietf.org/doc/html/rfc5545) iCalendar 解析、构建、校验与序列化库。

## 安装

```bash
go get github.com/saltfishpr/ical
```

## 快速开始

### 解析

```go
// 读取 .ics 文件
data, _ := os.ReadFile("event.ics")
cal, err := ical.ParseCalendar(data)
```

### 构建

```go
cal := ical.NewCalendar("//example.com//product")
event := ical.NewEvent("uid-1@example.com", time.Now()).
	AddDateTime(ical.PropDTStart, time.Now()).
	AddText(ical.PropSummary, "团队周会")
cal.AddComponent(event)

fmt.Println(cal) // 序列化为 RFC 5545 文本
```

### 校验

```go
errs := event.Validate()
for _, e := range errs {
    fmt.Println(e)
}
```

### 递归事件展开

```go
rulestr := "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;COUNT=10"
rule, err := ical.ParseRecurrenceRule(rulestr)

for dt := range rule.Expand(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)) {
    fmt.Println(dt)
}
```

## 核心 API

### 组件（Component）

| 构造函数                  | 用途      |
| ------------------------- | --------- |
| `NewCalendar(prodid)`     | VCALENDAR |
| `NewEvent(uid, stamp)`    | VEVENT    |
| `NewTodo(uid, stamp)`     | VTODO     |
| `NewJournal(uid, stamp)`  | VJOURNAL  |
| `NewFreeBusy(uid, stamp)` | VFREEBUSY |
| `NewTimezone(tzid)`       | VTIMEZONE |

### 属性访问

| 方法                                   | 对应 RFC 5545 类型 |
| -------------------------------------- | ------------------ |
| `AddText` / `Text`                     | TEXT               |
| `AddDateTime` / `DateTime`             | DATE-TIME          |
| `AddDate` / `Date`                     | DATE               |
| `AddTime` / `Time`                     | TIME               |
| `AddDuration` / `Duration`             | DURATION           |
| `AddInteger` / `Integer`               | INTEGER            |
| `AddFloat` / `Float`                   | FLOAT              |
| `AddBoolean` / `Boolean`               | BOOLEAN            |
| `AddGeo` / `Geo`                       | GEO                |
| `AddPeriod` / `Period`                 | PERIOD             |
| `AddURI` / `URI`                       | URI                |
| `AddBinary` / `Binary`                 | BINARY             |
| `AddCalAddress`                        | CAL-ADDRESS        |
| `AddUTCOffset` / `UTCOffset`           | UTC-OFFSET         |
| `AddRecurrenceRule` / `RecurrenceRule` | RRULE              |

### 值类型

提供完整的 Format/Parse 配对：

- `EncodeText` / `DecodeText`
- `FormatBoolean` / `ParseBoolean`
- `FormatInteger` / `ParseInteger`
- `FormatFloat` / `ParseFloat`
- `FormatDateTime` / `ParseDateTime`
- `FormatDuration` / `ParseDuration`
- `FormatUTCOffset` / `ParseUTCOffset`
- `FormatPeriod` / `ParsePeriod` / `ParsePeriodList`
- `FormatGeo` / `ParseGeo`
- `FormatCalAddress`
- `FormatBinary` / `ParseBinary`
- `Time` / `ParseTime`

### 低级 API

当类型化辅助方法不够用时：

- `AddProperty` — 手动添加任意属性
- `PropertiesByName` — 按名称检索所有匹配属性
- `AddComponent` — 添加子组件（用于嵌套结构）
- 未知和自定义 `X-*` 属性在解析 → 编辑 → 序列化的往返过程中完整保留

## 致谢

本项目灵感来源于 [github.com/collective/icalendar](https://github.com/collective/icalendar)。

## 许可证

MIT
