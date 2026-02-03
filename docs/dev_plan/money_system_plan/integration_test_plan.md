## Previous steps

1. Create an appointment

## Integration test flow

- Open cash session with 500 floating

1. Add charge on an appointment. Service 1, internal doctor -> 1500
2. Check balance = 1500
3. Add 2nd charge to an appointment. Service 2, external doctor -> 1500
4. Check balance = 3000
5. Add total correction for charge 2
6. Check balance = 1500
7. Add 2nd charge. Service 2, external doctor -> 1000
8. Check balance = 2500
9. Add first payment of 500, cash, mxn
10. Add second payment of 25, cash, usd, exchange rate 1 usd -> 20 mxn
11. check balance = 1500
12. Add third payment of 1000, card, mxn
13. check balance = 500

- Create charge on second appointment. Service 1, internal doctor -> 500
- Add payment, 500, card, mxn
- check balance = 0

## Get session details. Should show:

- The right entries
- the expected amounts:
  - cash, mx, 1000
  - cash usd 20
  - card mx 1000
  - card usd 0
- payments_summary: not entirely sure

## Link to document with test

- https://docs.google.com/spreadsheets/d/1N2PENxbF8JDsERqx2KGzwRm7xJtaCy9A5sd69K7WqXA/edit?gid=0#gid=0
