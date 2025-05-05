### zkLend Hack (February 2025): Detailed, Simplified Explanation

#### Overview

On February 11, 2025, zkLend, a crypto-based lending platform on StarkNet, suffered a significant hack. Attackers exploited specific software vulnerabilities, leading to losses of approximately \$9.6 million.

---

### Key Concepts Explained

#### What is zkLend?

zkLend is a decentralized lending platform. Users deposit crypto assets as collateral to borrow other assets.

#### Important Terms:

- **Collateral**: Assets deposited as security to borrow other assets.
- **Accumulator (lending_accumulator)**: A global number used to calculate the actual value of a user's deposited collateral based on market activities.
- **Flash Loans**: Very short-term crypto loans repaid within the same transaction.
- **Wei**: The smallest unit of Ethereum, analogous to cents in dollars (1 Ether = 10^18 Wei).

---

### How the Exploit Happened (Step-by-Step)

#### **Phase 1: Manipulating the Accumulator**

**Step 1: Exploiting an Empty Market**

- Before the attack, the market for the asset (wstETH) was empty.
- The attacker deposited a tiny amount (1 wei of wstETH).
  - 1 wei = 0.000000000000000001 wstETH (almost zero in value)
- This tiny deposit set the initial value of the accumulator to **1**.

**Step 2: Using Flash Loans to Inflate the Accumulator**

- The attacker took flash loans: borrowed small amounts but repaid significantly larger amounts.
  - Example: Borrowed **1 wei**, returned **1000 wei**.
  - Excess repayment (**999 wei**) treated as "donation" increasing the accumulator dramatically.
- Repeated this action 10 times, inflating the accumulator from **1** to approximately **4,069,297,906,051,644,020 (4.069 × 10¹⁸)**.

_Example Simplified:_

| Transaction   | Borrow | Repay | Excess ("Donation") | New Accumulator |
| ------------- | ------ | ----- | ------------------- | --------------- |
| Initial       | -      | -     | -                   | 1               |
| Flash loan 1  | 1 wei  | 1000  | 999 wei             | 851             |
| ...           | ...    | ...   | ...                 | ...             |
| Flash loan 10 | 1 wei  | 1000+ | 999+ wei            | 4.069×10¹⁸      |

---

#### **Phase 2: Exploiting Rounding Errors**

**Key Vulnerability:**

- The protocol rounds down when calculating amounts for deposits and withdrawals.
- Due to the accumulator's inflated value, these small rounding errors became significant.

**How it worked in practice:**

- Attacker's initial raw balance after manipulation was **1** (from 1 wei deposit).

**Steps of Exploitation:**

1. Deposited an amount slightly higher than current balance:
   - Example: Deposited **4.069 wstETH**, increasing raw balance from **1 to 2**.
2. Deposited more (**8.138 wstETH**), raw balance from **2 to 4**.
3. Immediately withdrew a calculated amount (**6.104 wstETH**):
   - Calculation: `6.104 / 4.069×10¹⁸ = ~1.5`
   - Rounded down to **1**, only reducing raw balance from **4 to 3** instead of the correct **2.5**.

_Simplified Visualization:_

| Action   | Amount       | Raw Balance (start) | Raw Balance (end) | Rounded Result       |
| -------- | ------------ | ------------------- | ----------------- | -------------------- |
| Deposit  | 4.069 wstETH | 1                   | 2                 | -                    |
| Deposit  | 8.138 wstETH | 2                   | 4                 | -                    |
| Withdraw | 6.104 wstETH | 4                   | **3**             | 1 (rounded from 1.5) |

- Each cycle increased the attacker’s balance slightly.
- Repeated multiple times, significantly inflating the collateral balance.

---

### **Extracting Funds**

- Using inflated collateral balance (\~7015 wstETH, worth millions), attackers borrowed assets:
  - ETH, USDC, STRK, USDT
- Then moved these stolen funds to other platforms.

**Stolen Amount Summary:**

- Ethereum (ETH): ~~2214 ETH (~~\$4.3 million)
- USD Coin (USDC): \~\$1.55 million
- StarkNet token (STRK): ~~7.4 million STRK (~~\$3.3 million)
- Tether (USDT): \~\$518,226

**Total Loss: \~\$9.57 million**

---

### Why Was This Possible?

The exploit combined three individually minor issues into a significant vulnerability:

1. Allowing tiny deposits to start a market.
2. Accepting "donations" through flash loan repayments.
3. Rounding down in withdrawal calculations, amplified by a hugely inflated accumulator.

---

### Visual Summary

```
1 wei Deposit → Tiny initial collateral
↓
Flash Loans (1 wei borrow, huge repay)
↓
Accumulator massively inflated
↓
Repeated Deposits & Withdrawals
↓
Exploited rounding errors
↓
Inflated collateral balance
↓
Borrowed and stole ~$9.57 million in other assets
```

---

### Aftermath and Response

- zkLend immediately suspended operations, stopping further losses.
- Security improvements were introduced, specifically addressing:
  - Market initialization
  - Donation mechanisms
  - Precision and rounding safeguards

---

### Lessons Learned for Software Engineers

- **Initialization Vulnerabilities:** Always validate initial market conditions.
- **Rounding Precision:** When precision matters financially, use precise math methods to avoid rounding exploitation.
- **Security by Design:** Every minor vulnerability can become critical in financial systems. Comprehensive threat modeling is essential.

---

This explanation clarifies exactly how the zkLend exploit unfolded, highlighting crucial security lessons for software engineers.
