# CALCULATE CCU

## Preconditions

- `wrk` (request and benchmark tool): `https://github.com/wg/wrk`

  Install on `MacOS`:

  ```
  brew install wrk
  ```

- Setup `redis`

  ```
  make init
  ```

- Start the server:

  ```
  make dev
  ```

## Calculate in time range per minute

- Mock requests to the server in through 2 minutes

  ```
  wrk -t12 -c400 -d2m http://127.0.0.1:1234
  ```

- Current number of ccu

  ```
  curl http://localhost:1234/ccu
  ```
