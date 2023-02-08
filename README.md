<div align="center"><img src="./docs/images/charonlogo.svg" /></div>
<h1 align="center">Charon<br/>Клиент связующего программного обеспечения распределнного валидатора</h1>

<p align="center"><a href="https://github.com/obolnetwork/charon/releases/"><img src="https://img.shields.io/github/tag/obolnetwork/charon.svg"></a>
<a href="https://github.com/ObolNetwork/charon/blob/main/LICENSE"><img src="https://img.shields.io/github/license/obolnetwork/charon.svg"></a>
<a href="https://godoc.org/github.com/obolnetwork/charon"><img src="https://godoc.org/github.com/obolnetwork/charon?status.svg"></a>
<a href="https://goreportcard.com/report/github.com/obolnetwork/charon"><img src="https://goreportcard.com/badge/github.com/obolnetwork/charon"></a>
<a href="https://github.com/ObolNetwork/charon/actions/workflows/golangci-lint.yml"><img src="https://github.com/obolnetwork/charon/workflows/golangci-lint/badge.svg"></a></p>

В этом репозитории содержится исходный код клиента распределенного валидатора _Charon_ (произносится 'харон'); это HTTP-клиент связующего программного обеспечения для организации стейкинга в сети Ethereum, позволяющий безопасно запускать одного валидатора, функционирование которого обеспечивается группой независимых узлов.

Charon предлагается к использованию вместе с web-приложением [Distributed Validator Launchpad] (https://goerli.launchpad.obol.tech/) для распределенного создания ключей валидаторов.

Charon может быть использован стейкерами для распределения ответственности между несколькими различными узлами и реализациями программных клиентов для запуска валидатора в сети Ethereum.

![Пример кластера Obol](./docs/images/DVCluster.png)

###### Кластер распределенного валидатора, использующий клиент Charon для устранения рисков отказа программного обеспечения клиентов и отказа оборудования

## Начало работы

Наиболее простым способом попробовать применение Charon является использование репозитория [charon-distributed-validator-cluster](https://github.com/ObolNetwork/charon-distributed-validator-cluster), содержащего конфигурацию docker compose для запуска полного кластера Charon на вашем устройстве.

## Документация

 Сайт [Obol Docs](https://docs.obol.tech/) является наилучшим местом для ознакомления с информацией. 
 Наиболее важные разделы на нем: [введение](https://docs.obol.tech/docs/intro), [основные понятия](https://docs.obol.tech/docs/int/key-concepts), [charon](https://docs.obol.tech/docs/dv/introducing-charon).

Для ознакомления с подробной документацией по этому репозиторию смотрите директорию [docs](docs):

Для ознакомления с подробной документацией по этому репозиторию смотрите директорию [docs](docs): 
- [Настройка](docs/configuration.md): Настройка узла Charon 
- [Структура](docs/architecture.md): Обзор архитектуры кластера и узла Charon 
- [Структура проекта](docs/structure.md): Структура директорий проекта
- [Модель ветвления и релизов](docs/branching.md): Модель ветвления и релизов Git 
- [Рекомендации для Go](docs/goguidelines.md): Рекомендации и принципы в отношении разработки на Go 
- [Участие](docs/contributing.md): Как внести свой вклад в создание Charon; Git-хуки, PR-шаблоны и т.п. 


Для ознакомления с документацией по исходному коду доступна документация пакета [charon godocs](https://pkg.go.dev/github.com/obolnetwork/charon). 

## Поддерживаемые клиенты для слоя консенсуса

Charon интегрируется в стек решений для уровня консенсуса Ethereum в качестве связующего программного обеспечения между клиентом валидатора и Beacon-узлом через официальный [Eth Beacon Node REST API](https://ethereum.github.io/beacon-APIs/#/). Charon поддерживает любые реализаци Beacon-узлов, в которых присутствует Beacon API. Charon стремится поддерживать любые обособленные реализации клиентов валидатора, использующих Beacon API.

| Клиент                                             | Beacon-узел | Клиент валидатора | Примечание                                       |
|----------------------------------------------------|-------------|-------------------|--------------------------------------------------|
| [Teku](https://github.com/ConsenSys/teku)          | ✅           | ✅                 | Поддерживается полностью                         |
| [Lighthouse](https://github.com/sigp/lighthouse)   | ✅           | ✅                 | Поддерживается полностью                         |
| [Lodestar](https://github.com/ChainSafe/lodestar)  | ✅           | *️⃣                 | Проблема в совместимости с DVT                   |
| [Vouch](https://github.com/attestantio/vouch)      | *️⃣           | ✅                 | Представлен только клиент валидатора             |
| [Prysm](https://github.com/prysmaticlabs/prysm)    | ✅           | 🛑                 | Клиент валидатора требует использования gRPC API |
| [Nimbus](https://github.com/status-im/nimbus-eth2) | ✅           | ✅                 | Поддержка ожидается скоро                        |

## Статус проекта

Obol Network все еще находится на ранних этапх и все его составляющие находятся в активной разработке. Изменения происходят быстро, поэтому регулярно следите за обновлениями для отслеживания развития.

Charon является распределенным валидатором, поэтому его основной ответственностью является выполнение задач для обеспечения функционирования валидатора. В следующей таблице демонстрируются клиенты и выполняемые ими задачи в публичной тестовой сети, а также, выполнение каких задач планируется к добавлению в них (🚧 )

| Задача \ Клиент                        |                      Teku                      |                    Lighthouse                    | Lodestar | Nimbus | Vouch | Prysm |
|--------------------------------------|:----------------------------------------------:|:------------------------------------------------:|:--------:|:------:|:-----:|:-----:|
| _Attestation_                        |                       ✅                        |                        ✅                         |    🚧    |   🚧   |  ✅   |  🚧   |
| _Attestation Aggregation_            |                       🚧                       |                        🚧                        |    🚧    |   🚧   |  🚧   |  🚧   |
| _Block Proposal_                     |                       ✅                        |                        ✅                         |    🚧    |   🚧   |  🚧   |  🚧   |
| _Blinded Block Proposal (mev-boost)_ | [✅](https://ropsten.beaconcha.in/block/555067) | [✅](https://ropsten.etherscan.io/block/12822070) |    🚧    |   🚧   |  🚧   |  🚧   |
| _Sync Committee Message_             |                       ✅                        |                        ✅                         |    🚧    |   🚧   |  🚧   |  🚧   |
| _Sync Committee Contribution_        |                       🚧                       |                        🚧                        |    🚧    |   🚧   |  🚧   |  🚧   |
