# Day 3 - demo

Un programma eseguibile per ogni slide del Giorno 3 con "demo" nelle note.
Nessuna dipendenza esterna. `go.mod` dichiara `go 1.27`.

| Slide                                   | Cartella            | Go minimo                   | Comando                                                                                         |
| --------------------------------------- | ------------------- | --------------------------- | ----------------------------------------------------------------------------------------------- |
| nil Channels Disable select Cases       | channels/nilselect  | 1.22                        | `go run ./channels/nilselect`                                                                   |
| close() as a Broadcast                  | channels/broadcast  | 1.22                        | `go run ./channels/broadcast`                                                                   |
| Buffered Channel as a Semaphore         | channels/semaphore  | 1.22                        | `go run ./channels/semaphore`                                                                   |
| Request/Response: a Channel of Channels | channels/chanofchan | 1.22                        | `go run -race ./channels/chanofchan`                                                            |
| Ownership Through Channel Direction     | channels/ownership  | 1.22                        | `go run ./channels/ownership`                                                                   |
| Priority, Timeouts and Tickers          | channels/priority   | 1.22                        | `go run ./channels/priority`                                                                    |
| Channel or Iterator?                    | channels/iterator   | 1.24                        | `go run ./channels/iterator`                                                                    |
| Finding Leaked Goroutines               | channels/leak       | 1.25 (test), 1.27 (profilo) | `go run ./channels/leak` · `go test -v ./channels/leak` · `SHOW_LEAK=1 go test ./channels/leak` |
| Language Changes                        | modern/lang         | 1.27                        | `go run ./modern/lang`                                                                          |
| Standard Library Highlights             | modern/stdlib       | 1.27                        | `go run ./modern/stdlib`                                                                        |

## Cosa mostra ogni demo

- **nilselect**: con i flag booleani il canale chiuso sveglia select milioni di volte; con `a = nil` solo 2. Più il pattern back-pressure con `send` nil.
- **broadcast**: un solo `close` sveglia tutti i worker; `Stop()` doppio senza panic; stesso meccanismo con `context.WithCancel`.
- **semaphore**: 20 dispositivi, mai più di 4 richieste contemporanee (verificato con un contatore atomico).
- **chanofchan**: 1000 incrementi da 10 goroutine senza mutex; un chiamante in timeout non blocca l'owner grazie alla reply bufferizzata.
- **ownership**: con `ctx.Done()` il produttore termina; senza, resta una goroutine bloccata (`runtime.NumGoroutine`).
- **priority**: su 1000 prove, senza priorità il publish passa prima dei comandi ~97% delle volte; con il primo select + default, 0.
- **iterator**: break su un canale lascia una goroutine bloccata, su un iteratore no; l'iteratore è ~7x più veloce.
- **leak**: `go run` stampa il profilo goroutineleak (Go 1.27); il test usa `testing/synctest` per verificare un timeout di un minuto in microsecondi; `SHOW_LEAK=1` mostra synctest che fallisce per la goroutine bloccata.
