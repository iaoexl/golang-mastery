# Go 语言从入门到精通（golang-mastery）

[![Docs: CC BY 4.0](https://img.shields.io/badge/Docs-CC%20BY%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by/4.0/deed.zh) [![Code: MIT](https://img.shields.io/badge/Code-MIT-blue.svg)](LICENSE)

> 一套把 Go 语言机理拆成可复现动作的开源知识库：板块1 至板块12 每篇都配齐「核心机理 → 操作清单 → 场景笔记法 → 案例小故事 → 实用验证 → 常见误区」，板块0 单列总纲。读者读完不算学会；命令跑出可判定的结果，读者才算过关。
>
> 本仓库不是语法速查表，也不是示例代码合集，而是一套**以验证通过为交付标准的学习工程**。全库分三条线：`docs/` 放知识库正文与导览，`cmd/`、`internal/`、`scripts/` 放能跑通全部判分表的可运行工程，`跑通记录.md` 留一次实跑的现场证据。三线互相指向，读者不必从零猜结论，也不能只抄结论。

## 仓库里有什么

| 路径 | 内容 | 取用方式 |
| --- | --- | --- |
| `docs/Go语言从入门到精通.md` | 导览加板块0 至板块12 的合订正文（即完整版） | 通读或按关键词全文检索 |
| `docs/分板块/README.md` | 8 节导览：板块地图、三十分钟起步、按症状反查、概念速查、实战映射、三级刻度等 | 不知道该翻哪一篇时先读它 |
| `docs/分板块/[0-9][0-9]_板块*.md` | 13 个独立文件：两位序号按 00 至 12 排序，板块0 总纲 + 板块1 至板块12 正课 | 按阶段单篇精读 |
| `internal/core/*.go` | 解析、切片、存储、并发、重试等机理的可测实现 | 改一行就跑 `go test -race ./internal/core` |
| `cmd/app/` | 带启动/存活/就绪三分探针与注入版本的示例服务，含 `httptest` 契约断言 | `go run ./cmd/app` |
| `scripts/` | 环境快照、四层门禁、回滚演练、空目录起步四条脚本 | 换机器后依次重跑 |
| `跑通记录.md` | 一份已填好的现场判定与三张场景卡 | 当对照模板用，别抄判定列 |
| `go.mod` | 模块路径 `example.com/atlas`，只用标准库、无第三方依赖 | `go mod verify` |

## 协议（文档 CC BY 4.0 · 代码 MIT）

### 协议摘要

> 本仓库的**整理汇编成果**（`docs/`、`跑通记录.md`、本 README 等文档）采用 **CC BY 4.0（知识共享署名 4.0 国际许可协议）**：任何人都可复制、发行、改编，也可用于商业目的；分发或二次加工时须标注本仓库出处并附上协议链接。
>
> **可运行源码与脚本**（`cmd/`、`internal/`、`scripts/`、`go.mod`）采用 **MIT**。Creative Commons 官方不建议用 CC 许可发布软件本身，只建议把 CC 许可用于软件文档；代码改用软件许可，使用者才不会遇到与主流开源许可不兼容的问题。
>
> 本仓库全部文件由 AI 工具依据公开资料生成、归纳与结构化。贡献者只对本仓库依法可能产生权利的整理汇编部分作出上述许可；纯机器生成、依法不受著作权保护的部分，贡献者不主张权利，任何人可自由使用且无需署名。所引 Go 官方文档、标准库文档、术语与命令示例的权利归各自权利人所有，不在本协议覆盖范围内。内容按「现状」提供，贡献者不保证准确性、不提供任何担保；商业使用前请自行核验第三方权利与合规要求。

### 完整协议

[CC BY 4.0 知识共享署名 4.0 国际许可协议](https://creativecommons.org/licenses/by/4.0/deed.zh) · [MIT License](https://opensource.org/license/mit) · 分层范围、AI 生成声明与 MIT 全文见仓库根目录 [LICENSE](LICENSE)

## 从这里开始

1. 先读 `docs/分板块/README.md`：它给出板块地图与按症状反查入口。
2. 在仓库根执行 `bash scripts/baseline.sh` 确认工具链与环境落点。
3. 执行 `bash scripts/check.sh`，末行出现 `ALL_GATES_PASSED` 后再按判分结果选板块精读。

## 跑法

```bash
bash scripts/baseline.sh                     # 换环境先做归因：环境类 / 依赖类 / 代码类
bash scripts/check.sh                        # 四层门禁，输出 ALL_GATES_PASSED 才算通过
bash scripts/rollback_drill.sh                # 回滚演练，输出 ROLLBACK_DRILL_PASSED 才算通过
bash scripts/bootstrap.sh ./my-new-project    # 空目录起步，输出 BOOTSTRAP_PASSED 才算通过
```

在仓库根目录执行这四条命令。前三条脚本会自行切换到工程根；`bootstrap.sh` 在传入的目标目录里工作，会**清空并重建**该目录，读者不要指向存有资料的目录。脚本只要求 `go` 在 `PATH` 中；工具链位于自定义目录时，先 `export GOROOT=...` 并把 `$GOROOT/bin` 加入 `PATH`。`rollback_drill.sh` 还需要 `curl`。`check.sh` 刻意关闭模块代理（`GOPROXY=off`）并开启 `-mod=mod`，让判定不依赖本机缓存命中。

## 从零起步与目录安排

`scripts/bootstrap.sh` 从一个空目录开始，把「写第一个文件 → 建文件夹 → 第一次跑通 → 打包发布」做成可重跑的命令。它不依赖本工程已有的文件，换台机器也能独立验板块1。脚本在目标目录内生成 `目录安排.md`、`发布清单.md` 与 `dist/` 产物（这三个路径不在本仓库中，由脚本现场创建）。

| 起步目录 | 解决什么 |
| --- | --- |
| `cmd/app/` | 放可执行入口，未来多入口各占一个子目录 |
| `internal/core/` | 放只被本模块内导入的实现 |
| `scripts/` | 放门禁、快照、演练与起步脚本 |
| `dist/`（不入库） | 放产物、元数据摘要与发布清单 |

## 目录与板块对应

| 路径 | 承载的知识库板块 | 可验证的机理 |
| --- | --- | --- |
| `go.mod` | 1_1、6_1 | 模块边界、最小版本选择、可复现构建 |
| `internal/core/parse.go` | 2_1 | 零值、`%w` 包装链、`errors.Join` 错误聚合 |
| `internal/core/slice.go` | 3_1 | 切片描述符共享、完整切片表达式钉死 cap |
| `internal/core/store.go` | 4_1、5_1 | 消费方定义窄接口、编译期断言、类型集合、锁护临界区 |
| `internal/core/worker.go` | 5_1 | goroutine 退出条件表、channel 关闭规则、`WaitGroup` 收敛 |
| `internal/core/retry.go` | 10_1 | 退避抖动、错误分类、幂等键、总耗时预算 |
| `internal/core/core_test.go` | 8_1 | 表驱动、`t.Helper`、`t.Cleanup`、`t.Parallel`、`t.TempDir`、`t.Setenv` |
| `internal/core/bench_test.go` | 8_1、9_1 | `ResetTimer`、`ReportAllocs`、结果消费、模糊种子语料 |
| `cmd/app/main.go` | 7_1、10_1 | 结构化日志、探针分类、注入版本、优雅关闭 |
| `cmd/app/main_test.go` | 7_5、10_5 | `httptest` 契约断言、在途请求收敛 |
| `scripts/check.sh` | 8_5、12_5 | 四层门禁与逐项判分 |
| `scripts/rollback_drill.sh` | 12_1、12_5 | 双版本切换、探针恢复、幂等重放与数据无损 |
| `scripts/baseline.sh` | 1_3 | 环境快照卡的原始输出 |

## 红—绿对照

`internal/core/core_test.go` 里每条断言都能被主动破坏。读者可逐一试这五处改动：删掉 `Snapshot` 里的 `copy`；把 `Window` 的完整切片表达式改成普通切片；去掉 `FetchWithRetry` 的错误分类；删掉 `Collect` 发送方的提前取消检查（`ctx.Err()` 判定）；把 `parse.go` 里的 `%w` 改成 `%v` 让错误链断掉。

五处改动各自对应一条会变红的断言。删了实现仍全绿的断言属于假断言，读者应先修断言再继续。

另说明一处容易误判的地方：本工程只用标准库，没有第三方依赖，因此 `go mod tidy` 不会生成 `go.sum`，`go mod verify` 仍会输出全部校验通过。知识库 6_1 讲的 `go.sum` 是通用口径，接入外部模块后才会落地为文件；判分表看的是「版本库差异为零」，不是「必须存在某个文件」。

## 使用建议与边界声明

- 内容仅供学习与自我训练参考，**不构成生产环境可直接上线的代码、安全承诺或任何专业意见**。不同 Go 版本在提示文案与默认参数上有差异，读者应以自己机器上实际跑出的结果为准，遇到不一致时查 `go doc` 与本地输出。
- `跑通记录.md` 是对照模板，不是标准答案。换机器重跑应另起记录，不预填、不补签；把别人的「通过」抄进自己的清单，等于把 12_5 的一票否决架空。
- 所引官方文档、标准库文档与工具的著作权归原作者与权利人所有；本仓库仅对**整理汇编成果**采用 CC BY 4.0，可运行源码与脚本采用 MIT。
- AI 生成内容的权利与免责说明见上文「协议摘要」，此处不再重复；二创请注明出处，商业使用前请自行确认合规。
