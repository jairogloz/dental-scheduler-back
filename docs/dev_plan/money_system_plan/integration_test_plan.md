## Previous steps

1. Create an appointment

## Integration test flow

### 1) Cash session setup

1. Open cash session with starting float = 500 MXN.

### 2) Appointment 1 flow

1. Create appointment 1.
2. Add charge #1: service 1, internal doctor, commission 30%, amount = 1500 MXN.
3. Assert appointment 1 balance = 1500.
4. Add charge #2: service 2, external doctor, doctor fee 500 MXN, amount = 1500 MXN.
5. Assert appointment 1 balance = 3000.
6. Add full correction for charge #2 (rollback full 1500 MXN).
7. Assert appointment 1 balance = 1500.
8. Add charge #3: service 2, external doctor, doctor fee 500 MXN, amount = 1000 MXN.
9. Assert appointment 1 balance = 2500.
10. Add payment #1: 500 MXN, cash.
11. Add payment #2: 25 USD, cash, exchange rate 1 USD = 20 MXN.
12. Assert appointment 1 balance = 1500.
13. Add payment #3: 1000 MXN, card.
14. Assert appointment 1 balance = 500.

### 3) Appointment 2 flow

1. Create appointment 2.
2. Add charge: service 1, internal doctor, commission 30%, amount = 500 MXN.
3. Add payment: 500 MXN, card.
4. Assert appointment 2 balance = 0.

### 4) Session details assertions

Assert cash session details include the expected entries and raw currency buckets:

- cash MXN = 1500 (500 opening float + 500 payment #1 + 500 MXN equivalent from USD payment)
- cash USD = 25
- card MXN = 1500 (1000 from appointment 1 + 500 from appointment 2)
- card USD = 0

For now, validate raw currency buckets only (no converted aggregate assertions).

## Link to document with test

- https://docs.google.com/spreadsheets/d/1N2PENxbF8JDsERqx2KGzwRm7xJtaCy9A5sd69K7WqXA/edit?gid=0#gid=0
