<h1 align="center">go-itmo</h1>

<p align="center"><strong>Неофициальная Go-библиотека для сервисов Университета ИТМО:
<a href="https://my.itmo.ru">my.itmo</a>, <a href="https://bars.itmo.ru">БАРС</a> и ITMO.ID</strong></p>

<p align="center">
<a href="https://pkg.go.dev/github.com/kewldan/go-itmo"><img src="https://pkg.go.dev/badge/github.com/kewldan/go-itmo.svg" alt="Go Reference"></a>
<a href="https://github.com/kewldan/go-itmo/actions/workflows/ci.yml"><img src="https://github.com/kewldan/go-itmo/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
<a href="https://goreportcard.com/report/github.com/kewldan/go-itmo"><img src="https://goreportcard.com/badge/github.com/kewldan/go-itmo" alt="Go Report Card"></a>
<a href="LICENSE"><img src="https://img.shields.io/github/license/kewldan/go-itmo" alt="License"></a>
</p>

**Покрытие:** все 452 маршрута личного кабинета my.itmo.ru
(809 методов, включая служебные разделы), все 81 маршрут REST API БАРС
(91 метод) и полный цикл входа ITMO.ID.

## Установка

```bash
go get github.com/kewldan/go-itmo@latest
```

Нужен Go 1.27+. Зависимости: только `golang.org/x/oauth2` и `golang.org/x/net`.

## Пакеты

| Пакет | Что делает |
|---|---|
| [`itmoid`](itmoid) | Вход в ITMO.ID (Keycloak): логин/пароль, одноразовый код (OTP), SSO без пароля, PKCE, обновление и хранение токенов, выход, userinfo, разбор ID-токена, impersonation |
| [`myitmo`](myitmo) | Клиент my.itmo.ru: все разделы кабинета — студенческие и служебные — плюс QR-пропуск `qr.itmo.su` |
| [`bars`](bars) | Клиент БАРС: собственная сессия, тихое продление через SSO, выбор периода, журналы, оценки, планы контрольных точек, отчёты |

## Быстрый старт

```go
ctx := context.Background()
auth := itmoid.New()

tok, err := auth.Login(ctx, itmoid.MyITMO, itmoid.Credentials{
	Username: os.Getenv("ITMO_LOGIN"),
	Password: os.Getenv("ITMO_PASSWORD"),
})
if err != nil {
	log.Fatal(err)
}

// ITMO.ID ротирует refresh token при каждом обновлении — сохраняйте каждый новый.
save := func(t *oauth2.Token) { _ = os.WriteFile(".refresh-token", []byte(t.RefreshToken), 0o600) }
save(tok)

client := myitmo.New(auth.TokenSource(ctx, itmoid.MyITMO, tok, save))

days, err := client.Schedule.Personal(ctx, myitmo.Today(), myitmo.Today().AddDays(7))
```

## Аутентификация

### Логин и пароль

`Authenticator.Login` проходит форму ITMO.ID так же, как браузер: получает
страницу входа, отправляет логин и пароль, забирает код из редиректа и
обменивает его на токены (для my.itmo — с PKCE). Логин и пароль не хранятся.

Ошибки различимы через `errors.Is`:

| Ошибка | Когда |
|---|---|
| `itmoid.ErrInvalidCredentials` | неверный логин/пароль или OTP; текст с формы — в `*itmoid.FormError` |
| `itmoid.ErrOTPRequired` | включена двухфакторка, а `Credentials.OTP` не задан |
| `itmoid.ErrLoginRequired` | вход без пароля (SSO), но сессии ITMO.ID нет |
| `itmoid.ErrUnexpectedPage` | ITMO.ID показал страницу, которую нужно пройти в браузере (смена пароля и т. п.) |

### Двухфакторная аутентификация

```go
tok, err := auth.Login(ctx, itmoid.MyITMO, itmoid.Credentials{
	Username: login, Password: password,
	OTP: func(ctx context.Context) (string, error) { return askUser("Код из приложения") },
})
```

### Refresh token

```go
ts := auth.FromRefreshToken(ctx, itmoid.MyITMO, storedRefreshToken, saveToken)
client := myitmo.New(ts)
```

`itmoid.RefreshExpiry(tok)` говорит, когда истечёт refresh token;
`itmoid.IDClaims(tok)` возвращает ISU, ФИО и фото из ID-токена (без проверки
подписи — только для отображения).

### Вход на странице ITMO.ID (WebView, браузер)

Если пароль не должен попадать в приложение, покажите пользователю настоящую
страницу ITMO.ID и перехватите редирект на callback:

```go
state, verifier := itmoid.NewState(), oauth2.GenerateVerifier()
loginURL := auth.AuthCodeURL(itmoid.MyITMO, state, verifier) // открыть в WebView

// WebView перешёл на https://my.itmo.ru/login/callback?state=...&code=...
code, err := itmoid.ParseCallback(itmoid.MyITMO, auth.Issuer(), redirectURL, state)
tok, err := auth.Exchange(ctx, itmoid.MyITMO, &itmoid.Code{Value: code, Verifier: verifier, State: state})
```

`ParseCallback` строго проверяет схему, хост, путь, `state`, `iss`, отсутствие
`error`, фрагмента и повторяющихся параметров. Для БАРС используйте
`itmoid.BARS` (без PKCE, `verifier` пустой) и передайте код в `bars.Client.Login`.

### Готовый access token

Например, скопированный из браузера. Он живёт несколько минут и сам не обновляется:

```go
client := myitmo.New(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}))
```

### Прочее

| Метод | Что делает |
|---|---|
| `auth.Refresh(ctx, app, refreshToken)` | принудительно обновить токены |
| `auth.Logout(ctx, app, refreshToken)` | завершить SSO-сессию ITMO.ID |
| `auth.UserInfo(ctx, ts)` | OIDC userinfo: ISU, ФИО, почта |
| `itmoid.IDClaims(tok)` | те же данные из ID-токена без запроса (подпись не проверяется) |
| `itmoid.RefreshExpiry(tok)` | когда истечёт refresh token |
| `auth.Impersonate(ctx, app, accessToken, subject)` | войти от имени другого пользователя (RFC 8693); только для аккаунтов поддержки |
| `itmoid.MyITMODev` | стенд `dev.my.itmo.su` (вместе с `myitmo.WithBaseURL(myitmo.DevBaseURL)`) |

### Один вход — несколько сервисов

`Authenticator` хранит SSO-cookie ITMO.ID. После входа в my.itmo тот же
`Authenticator` получает код для БАРС без пароля:

```go
barsClient := bars.New(bars.WithAuthenticator(auth))
err := barsClient.LoginSSO(ctx)
```

## my.itmo

Сервисы сгруппированы так же, как разделы кабинета:

```go
client.Schedule.Personal(ctx, from, to)              // расписание
client.RecordBook.Specializations(ctx)               // зачётка
client.Sport.Schedule(ctx, myitmo.SportScheduleParams{DateStart: from, DateEnd: to}) // спорт
client.Personalities.SearchAll(ctx, "Иванов", 50)    // iter.Seq2 по всем страницам
client.QR.Pass(ctx)                                  // QR-пропуск
```

| Поле клиента | Раздел кабинета | Маршруты | Методов | Кому |
|---|---|---|---:|---|
| `Schedule` | Расписание | `/api/schedule` | 3 | все |
| `RecordBook` | Зачётка | `/api/record_book` | 3 | студент |
| `StudyPlan` | Учебный план, выбор и замена дисциплин | `/api/eduPlanNew`, `/api/eduPlan` | 9 | студент |
| `IndividualPlan` | Индивидуальные планы | `/api/individual-plan` | 35 | сотрудник |
| `Election` | Выборность дисциплин и потоков | `/api/election` | 15 | студент |
| `Intro`, `Facultative` | Вводные курсы и факультативы | `/api/intro`, `/api/facultative` | 13 | студент |
| `Sport` | Физкультура: запись, баллы, секции, соревнования | `/api/sport` | 34 | студент |
| `Personalities` | Персоналии | `/api/personalities` | 4 | все |
| `System` | Главный экран, меню, сервисы | `/api/system` | 3 | все |
| `Requests` | Заявки (классические) | `/api/requests` | 12 | все |
| `RequestsV2` | Заявки на процессах, конструктор форм и процессов | `/api/requests/v2` | 58 | все / админ |
| `Agreements` | Согласия и ознакомления | `/api/agreements` | 4 | все |
| `Sign` | Электронная подпись (простая, Контур, Госключ) | `/api/sign` | 11 | все |
| `Assistant` | ИИ-ассистент (SSE-стрим) | `/api/assistant` | 7 | все |
| `Finances` | Стипендии, оплата обучения, зарплата | `/api/finances` | 13 | студент / сотрудник |
| `Dormitory` | Общежитие: заселение, договоры, оплата | `/api/dormitory` | 11 | студент |
| `DMS` | ДМС | `/api/dms` | 13 | все / HR |
| `Booking` | Бронирование аудиторий | `/api/booking` | 12 | все |
| `Navigator` | Навигатор по корпусам | `/api/navigator` | 3 | все |
| `Queues` | Электронные очереди | `/api/queues` | 7 | все |
| `Practices` | Практики и отчёты | `/api/practices` | 22 | студент |
| `Employments` | Трудоустройство (HR) | `/api/employments` | 31 | сотрудник |
| `GIAStudents` | ГИА и ВКР, приложение к диплому | `/api/gia-students` | 50 | студент |
| `GIA` | ГИА: комиссии, защиты, рецензенты, приказы | `/api/gia` | 175 | сотрудник |
| `Adviser` | Тьютор: карточка студента | `/api/services/adviser` | 15 | сотрудник |
| `Checklist` | Обходной лист | `/api/services/checklist` | 1 | студент |
| `Constructor` | Конструктор РПД/РПП/ГИА | `/api/constructor` | 102 | сотрудник |
| `ConstructorEP` | Конструктор ОП, банк модулей, КУГ | `/api/constructor-ep` | 112 | сотрудник |
| `Vacation` | Отпуска | `/api/vacation` | 30 | сотрудник |
| `QR` | QR-пропуск | `qr.itmo.su` | 1 | все |

Методы служебных разделов (ГИА, конструктор ОП, отпуска, трудоустройство,
администрирование заявок) доступны только аккаунтам с соответствующими правами;
иначе сервер отвечает 403 (`(*myitmo.Error).IsForbidden()`).

### Ошибки

Любая ошибка сервера — `*myitmo.Error`: HTTP-статус, `error_code` из ответа
(my.itmo нередко отвечает HTTP 200 с ненулевым кодом) и локализованное
сообщение, которое можно показать пользователю.

```go
_, err := client.Sport.SignIn(ctx, lessonID)
var apiErr *myitmo.Error
if errors.As(err, &apiErr) && apiErr.Code == myitmo.SportErrorCannotSignIn {
	fmt.Println(apiErr.Message) // «нельзя записать студента: [...]»
}
```

`*myitmo.DecodeError` означает, что ответ не совпал с моделью — скорее всего,
сервер изменил тип поля. Пожалуйста, заведите issue.

### Непокрытое и сырые запросы

Если нужного метода ещё нет, `client.Do` отправит любой запрос с токеном и
разберёт стандартную обёртку `{error_code, error_message, result}`:

```go
var out []myThing
err := client.Do(ctx, http.MethodGet, "api/some/new/route", nil, nil, &out)
```

## БАРС

```go
client := bars.New(bars.WithAuthenticator(auth), bars.WithStore(myStore))
err := client.LoginPassword(ctx, itmoid.Credentials{Username: login, Password: password})

err = client.WithPeriod(ctx, "2025/2026", bars.Spring, func(ctx context.Context) error {
	disciplines, err := client.Disciplines(ctx, true)
	// ...
	journal, err := client.StudentJournal(ctx, planID, flow.Type, flow.Identifier)
	// ...
})
```

Способы входа в БАРС:

| Метод | Когда |
|---|---|
| `LoginPassword(ctx, creds)` | логин и пароль ITMO.ID (нужен `WithAuthenticator`) |
| `LoginSSO(ctx)` | без пароля, если тот же `Authenticator` уже входил (например, в my.itmo) |
| `Login(ctx, code)` | код ITMO.ID, полученный сами (WebView, см. выше) |
| `LoginParent(ctx, login, password)` | родительский аккаунт БАРС (не ITMO.ID) |
| `RegisterParent(ctx, token, login, password)` | регистрация родителя по ссылке-приглашению студента |
| `Impersonate(ctx, personalNumber)` | суперпользователь: работать от имени другого пользователя |
| `WithStore(store)` | сохранить сессию между запусками (по умолчанию — в памяти) |

- Сессия БАРС — заголовок `Bearer ...`, живёт ~30 минут и не обновляется.
  С `WithAuthenticator` (или `WithCodeSource`) клиент на HTTP 401 один раз
  получает новый код через SSO и повторяет запрос.
- Каталоги и журналы читаются в контексте периода, сохранённого на сервере и
  общего для всех сессий пользователя. `WithPeriod` переключает период и не даёт другим
  вызовам `WithPeriod` этого клиента сменить его, пока выполняется функция.
- `StudentMarks.Total` у пустого журнала равен 0 — проверяйте `HasAnyMark()`.
- `Approval.GradeCode()` переводит «Удвл., E» в формат my.itmo «3/E».
- Идентификаторы БАРС не совпадают с `discipline_id` и `est_id` my.itmo.

## QR-пропуск

`client.QR.Pass(ctx)` возвращает строку `Hex`. QR-код кодирует сам этот текст
(byte mode, уровень коррекции L), например с `github.com/skip2/go-qrcode` или
`rsc.io/qr`.

## Безопасность

- Токены отправляются только на хосты my.itmo и qr.itmo.su, даже если в `Do`
  передан абсолютный URL на другой хост.
- Логирование (`WithLogger`) пишет только метод, путь, статус и время ответа —
  без токенов, query-параметров и тел.
- Ошибки БАРС не содержат тела ответа.
- Не храните пароль. Храните refresh token так же бережно, как пароль
  (`0600`, системное хранилище секретов).

## Модели

Поля, форма которых ещё не подтверждена, имеют тип `RawJSON` — их можно
разобрать самостоятельно. Если ответ сервера не совпал с моделью, придёт
`*myitmo.DecodeError`; пришлите issue или PR с обезличенным примером.

## Разработка

```bash
make build   # go build + go vet
make test    # go test -race ./...
make lint    # golangci-lint v2
```

Правила для контрибьюторов (и агентов) — в [AGENTS.md](AGENTS.md).

## CI

- **CI** (`ci.yml`): gofmt и `go mod tidy`, сборка, `go vet`, тесты с `-race`
  на Linux, macOS и Windows, отчёт о покрытии, golangci-lint, govulncheck.
- **CodeQL** (`codeql.yml`): статический анализ безопасности, также раз в неделю.
- **Release** (`release.yml`): на тег `vX.Y.Z` — тесты, GitHub Release с
  автоматическими заметками и регистрация версии в proxy.golang.org.
- **Dependabot**: еженедельные обновления Go-модулей и GitHub Actions.

## Лицензия

[MIT](LICENSE)

## Дисклеймер

Проект не связан с Университетом ИТМО. Используйте только со своей учётной
записью и в рамках правил университета; служебные методы — только при наличии
соответствующих прав.
