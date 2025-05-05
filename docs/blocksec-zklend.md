Title: zkLend Exploit Post-Mortem: Unraveling the Details and Clarifying Misunderstandings of the $10M Flash Loan Attack

URL Source: https://blocksec.com/blog/zklend-exploit-post-mortem-unraveling-the-details-and-clarifying-misunderstandings-of-the-10m-flash-loan-attack

Published Time: 2025-02-20

Markdown Content:
On February 12, 2025, zkLend \[1\], a lending protocol on **StarkNet**, was exploited for approximately $10M through a sophisticated manipulation of its accumulator mechanism. The attacker leveraged flash loans and rounding vulnerabilities to artificially inflating collateral values, borrowing **other assets** from the protocol to profit.

However, there remains a lack of detailed and accurate technical analysis from a security perspective. Despite existing analyses by other security researchers, which provided valuable insights, some misunderstandings persist—particularly regarding the attack analysis. zkLend’s later publication of the official post-mortem \[2\] offers a simplified description but lacks a detailed technical analysis. In this blog, we aim to provide a comprehensive examination to clarify the incident.

Key Takeaways (TL;DR)
---------------------

*   **The root cause of this incident stems from the combination of the following three issues**:
    
    *   _**The empty market initialization**_ allows arbitrary asset deposits.
    *   _**The specific donation mechanism in zkLend’s flash loan**_ enables manipulation of the accumulator, a global variable as a scaling factor to dynamically adjust users’ collateral balances.
    *   _**Precision loss occurs due to truncation**_. Unlike classical precision loss in division, the denominator starts at 1 but was inflated to a very large value, causing underestimation during the burning of the share token.
*   **The attacker did not profit from wstETH deposited by other users**. Instead, the attacker leveraged the vulnerabilities to manipulate the collateral balance, using a small amount of wstETH as the initial capital to increase the collateral balance up to over 7,000 wstETH, thereby enabling the borrowing of other assets from the market.
    

In the forthcoming sections, we will first offer some crucial background information about zkLend. Subsequently, we will conduct an in-depth analysis of the issues and the associated attack.

0x1 Background: Understanding zkLend’s Core Protocol
----------------------------------------------------

zkLend is a lending project on StarkNet that supports common lending protocols such as collateralized loans and flash loans. Let’s dive into the implementation details of these two protocols.

### 0x1.1 Collateralized Loans

A collateralized loan refers to the process where users deposit specific assets into the protocol as collateral in exchange for borrowing other assets. The value of the collateral is used to determine the borrowing capacity. It’s important to note that lending protocols typically don’t store the collateral’s asset value directly; instead, they calculate it using the formula:

> collateral\_balance = lending\_accumulator \* raw\_balance

Specifically, the `lending_accumulator` is a scaling factor that dynamically adjusts each user’s collateral value, while `raw_balance` represents the actual share the user holds in the market. `raw_balance` is derived from the `collateral_balance` using the `lending_accumulator`.

_**What is purpose of this design?**_ It enables the protocol to efficiently manage collateral value while incentivizing users to deposit assets. By allocating a portion of the protocol’s earnings to collateral providers, the `lending_accumulator` increases, thereby amplifying the value of all users’ collateral proportionally and simultaneously.

### 0x1.2 Flash Loans on zkLend

A flash loan is a type of uncollateralized loan where users can borrow assets from the protocol for a very short period, typically within a single transaction. If the borrower fails to repay the loan or meet the specified conditions, the entire transaction is reverted, and the loan is not executed.

In zkLend's flash loan implementation, there is a unique _**donation**_ mechanism. Specifically, when users repay assets, they not only return the required minimum amount but can also contribute extra funds as a donation. The protocol tracks these donated funds and updates the `lending_accumulator` accordingly. This process is implemented in `thesettle_extra_reserve_balance()` function. The formula for updating the `lending_accumulator` is as follows:

> new\_accumulator = (reserve\_balance + totaldebt - amount\_to\_treasury) / ztoken\_supply

*   `reserve_balance`: The total amount of underlying token (e.g., wstETH) held in the contract, which includes the amount of tokens donated by users.
*   `totaldebt`: The total debt of all borrowing users.
*   `amount_to_treasury`: The amount of protocol’s revenue.
*   `ztoken_supply`: The total supply of the share token (e.g. zwstETH). When users deposit wstETH, the zkLend ztoken contract mints an equivalent amount of zwstETH.

Having understood the core protocol of zkLend, we will now formally explain how the attacker manipulated their collateral assets by manipulating the `lending_accumulator` and `raw_balance` variables.

0x2 Attack Analysis
-------------------

The attacker exploited the following mechanisms and vulnerabilities in the zkLend contract to manipulate the value of the collateral:

*   **Manipulation of `lending_accumulator`**
    *   **Empty market**: Before the attack, the zkLend market for wstETH tokens was empty, providing the perfect condition for manipulation. Moreover, the zkLend Market contract allows anyone to deposit any amount of assets into an empty market. The attacker deposited a small amount of assets to significantly inflate the `lending_accumulator` value.
    *   **Donation mechanism**: The zkLend Market contract’s `flash_loan()` function features a unique _**donation**_ mechanism. Specifically, when a user repays a flash loan, the Market contract calculates the excess funds returned and increases the global `lending_accumulator` variable, thereby amplifying the collateral values for all users in the contract.
*   **Manipulation of `raw_balance`**
    *   **Rounding behavior**: The division operation during the share token burning process uses truncation, which leads to an underestimation of the change in the user’s `raw_balance` during withdrawals.

By manipulating both of these variables, the attacker was able to increase the collateral balance to over _7,000_ wstETH and borrow _other assets_ from the market for profit.

### 0x2.1 Manipulating the lending\_accumulator Variable

#### 0x2.1.1 Empty Market Initialization

By examining the [transaction record](https://voyager.online/tx/0x01c0f62dc1c0c013fff273b8d25c2e761364c036e327285c3838b9415500d2f1) of the Market contract prior to the attack, we can observe that the attacker initially deposits **1 wei** of wstETH into the wstETH Market contract. By reviewing the internal calls of this transaction, it is evident that the wstETH Market contract held **0** wstETH, and the total supply of zwstETH was also **0**.

![Image 1](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/tx_0x01c0_7a31cd6b6c.png)

Therefore, we can confirm that there were no prior deposits or borrows in the zkLend wstETH market. Both the reserve\_balance and `ztoken_supply` were at their initial values of 0, and the initial value of the `lending_accumulator` was 1. This empty market scenario created the conditions for the subsequent attack, allowing the attacker to significantly amplify the `lending_accumulator` with a minimal amount of wstETH.

#### 0x2.1.2 Manipulating lending\_accumulator via Flash Loan

Next, in [this transaction](https://voyager.online/tx/0x039b6587b9d545cfde7c0f6646085ab0c39cc34e15c665613c30f148b569687c#overview), the attacker calls the `flash_loan()` function, borrowing **1 wei** wstETH and repaying **1000 wei** wstETH. The excess **999 wei** is treated as a **donation** and recorded into the contract’s `reserve_balance`.

![Image 2](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/tx_0x039b_1_e9539fae5e.png)

According to the formula for calculating the `lending_accumulator`, this transaction causes the `lending_accumulator` to increase from **1** to **851.0**.

![Image 3](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/tx_0x039b_2_5be43f4da8.png)

#### 0x2.1.3 Repeated Execution of flash\_loan()

The attacker executes a total of 10 `flash_loan()` calls, each time borrowing only 1 wei of wstETH but repaying a larger amount. As a result, `lending_accumulator` escalates to an astronomical value of _4,069,297,906,051,644,020_ (4.069 × 10^18), which coincidentally aligns with the decimal precision of wstETH.

### 0x2.2 Manipulating the raw\_balance Variable

After manipulating the `lending_accumulator` to approximately 4.069 × 10^18, the attacker called the `deposit()` function of the Market contract with _4.069297906051644020_ wstETH. Based on the latest value of the `lending_accumulator`, the attack contract’s `raw_balance` became **2**.

#### 0x2.2.1 The First Transaction Manipulating raw\_balance

In [this transaction](https://voyager.online/tx/0x001e3c2ebb5b4eafc4800c178dcbd6aa36233d40733bc419d6bce47f8c48d6e6#overview), the attacker called the `callflashloandraaan()` function of the attack contract. Although this contract is not open source, based on the internal call trace, it can be speculated that the logic of this function includes a loop that performs the following actions:

*   **Deposit**: The attacker deposits a certain amount of wstETH into the market contract.
*   **Withdraw**: The attacker withdraws the specific amount of wstETH.

![Image 4](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/tx_0x001e_1_01a65046e2.png)

**Token Transfer Record Analysis**

It can be observed that the amount of wstETH the attacker deposits is always an integer multiple of the `lending_accumulator`, for example, 2 times the value (e.g., _8.13859_) of the `lending_accumulator`.

However, the amount of wstETH withdrawn is 1.5 times the value (e.g., _6.10394_) of the `lending_accumulator`.

![Image 5](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/tx_0x001e_2_fbed74384c.png)

Through calculations, we can determine that the amount of wstETH withdrawn exceeds the amount deposited. Why does this happen?

**Rounding Behavior**

By reviewing the implementation of the `deposit()` and `withdraw()` methods, we can see that these two methods involve the minting and burning of zwstETH, respectively. Here’s how this works:

![Image 6](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/func_mint_b6cd722ec2.png)\`mint()\` function in the Market contract

![Image 7](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/func_burn_0ad11e8ec9.png)\`burn()\` function in the Market contract

The `mint()` and `burn()` processes both include a _**scale down logic**_. The scale down logic involves integer division with _**floor rounding**_ (_**rounding down**_ to the nearest integer), which plays a key role in the exploit.

![Image 8](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/func_div_80055887da.png)

When the attacker burns a certain amount of zwstETH, the **scaled down logic** is applied. Due to the manipulated value of the `lending_accumulator` being exceptionally high (around\* 4,069,297,906,051,644,020)\*, this division causes the attacker’s `raw_balance` to decrease by only 1 unit, despite burning over 6 zwstETH.

The attacker’s `raw_balance` changes are summarized in the following table:

![Image 9](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/raw_balance_changes_f3015947a3.jpg)

We can observe that in this transaction, the attacker repeatedly executes the _**Deposit - Withdraw**_ logic, exploiting the precision loss during the `withdraw()` function, which results in an underestimation of the `raw_balance` difference. Ultimately, the user’s `raw_balance` increased from _**2**_ to _**3**_, gaining an additional unit.

#### 0x2.2.2 Subsequent Attack Process

Subsequent [attack transactions](https://voyager.online/contract/0x04d7191dc8eac499bac710dd368706e3ce76c9945da52535de770d06ce7d3b26#transactions?ps=100&p=2) followed the same pattern as the first attack: the attacker repeatedly cycles through _**Deposit - Withdraw**_ transactions to acquire wstETH.

The acquired wstETH is re-deposited back into the market, further increasing the `raw_balance`, causing the attacker’s collateral value to keep rising.

![Image 10](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/tx_0x04d7_8828c4f468.png)

**Example explanation**

We use the following [transaction](https://voyager.online/tx/0x24cd75cd7cf619718bb62df404ce6d53c17f601645ea882ceae7933d5b07b8a) as an illustration.

![Image 11](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/tx_0x24cd_1_7084d5777c.png)

*   A total of 30 deposits were made, with 4.069 wstETH deposited each time.
*   A total of 30 withdrawals were made, with 6.104 wstETH withdrawn each time.
*   After this cycle, the attacker successfully extracted 61.39 wstETH, according to the calculations.

Additionally, it is worth noting that between these attack transactions, several `increase()` methods were called. These methods were used to transfer a specific amount of wstETH from the attacker’s account to the attack contract, which then provided the funds for subsequent deposits into the Market contract.

![Image 12](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/tx_0x24cd_2_3708ee2e99.png)

These actions boost the value of `raw_balance`, allowing the attacker to continue increasing the collateral value. Eventually, the attacker’s `raw_balance` reached **1,724**, with a value of **7,015.4** _**wstETH**_, which was sufficient to borrow _**other assets**_ from the market.

0x3 Profit Analysis
-------------------

### 0x3.1 Borrow Other Types of Funds

After manipulating the value of collateral value, the attacker **borrowed other types of funds** from the market and proceeded with the following transactions (excerpt):

![Image 13](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/borrow_other_assets_0e1e618928.png)

### 0x3.2 Bridge the Borrowed Funds to Layer1

By inspecting the bridge transactions of the [attacker’s contract](https://voyager.online/contract/0x04d7191dc8eac499bac710dd368706e3ce76c9945da52535de770d06ce7d3b26#bridgeTxns?ps=100&p=1), it can be observed that the attacker bridged part of the borrowed funds to Layer 1.

![Image 14](https://blocksec-static-resources.s3.us-east-1.amazonaws.com/assets/frontend/blocksec-strapi-online/contract_0x04d7_c03e998205.png)

0x4 Conculsion
--------------

In summary, this attack on the zkLend protocol highlights several important implications for the design and security of decentralized lending protocols:

*   **Market Initialization and Asset Deposit Conditions**: The empty market at the start allowed the attacker to deposit a small amount of wstETH and manipulate the `lending_accumulator`, gaining leverage for the exploit. Ensuring a sufficient liquidity base or limiting asset donations in early market stages could help prevent similar attacks.
*   **Importance of Proper Accumulator Mechanisms**: The attacker exploited the donation mechanism in the `flash_loan()` function to manipulate the `lending_accumulator`, inflating collateral values across all users. Protocols with accumulator-based mechanisms should safeguard against easy manipulation of scaling factors.
*   **Rounding Behavior and Precision Loss**: A rounding issue during zwstETH token burns led to precision loss and underestimation of `raw_balance`, allowing the attacker to manipulate the `raw_balance`. Protocols should use higher precision or validation checks to prevent such exploits.

Once again, this incident underscores the importance of _**timely notifications**_ regarding initialization and operational status, as well as _**proactive threat prevention**_ to mitigate potential losses.

Reference
---------

\[1\] https://zklend.com/

\[2\] zkLend’s security incident post-mortem: https://drive.google.com/file/d/10i1dh\_J89tPPw7KRcmFIVM6iNrJZAyfi/view
