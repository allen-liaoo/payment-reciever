Payment reciever logic
- Wake up after lambda call
- Lookup DB to retrieve list of (derived) middleware wallets to check
  - For each wallet:
    - Check if funds are recieved (>= threshold?)
    - If recieved, estimate sum gas prices of financing to middleware wallet, and middleware to reciever wallet
      - If >= threshold, sleep
      - Else,
        - Send ETH from financing wallet to middleware wallet
        - Once it is done sending, send Token from middleware wallet to reciever wallet

```bash
go run .
```