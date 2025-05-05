#[starknet::contract]
mod MinimalVulnerableMarket {
    use starknet::{ContractAddress, get_caller_address};
    use starknet::contract_address::ContractAddressZeroable;

    // Storage variables
    #[storage]
    struct Storage {
        reserve_balance: u256,
        total_supply: u256,
        lending_accumulator: u256,
        user_balances: LegacyMap<ContractAddress, u256>,
    }

    const SCALE: u256 = 1000000000000000000; // 1e18

    #[constructor]
    fn constructor(ref self: ContractState) {
        self.lending_accumulator.write(SCALE); // Initialize to 1.0
    }

    // VULNERABILITY: No market activation, no minimum deposit
    #[external(v0)]
    fn deposit(ref self: ContractState, amount: u256) -> u256 {
        let caller = get_caller_address();

        // Update reserve balance
        let new_reserve = self.reserve_balance.read() + amount;
        self.reserve_balance.write(new_reserve);

        // Calculate shares to mint
        let shares = self._calculate_shares(amount);

        // Update user's balance and total supply
        let new_balance = self.user_balances.read(caller) + shares;
        self.user_balances.write(caller, new_balance);

        let new_supply = self.total_supply.read() + shares;
        self.total_supply.write(new_supply);

        shares
    }

    #[external(v0)]
    fn flash_loan(ref self: ContractState, amount: u256, repay_amount: u256) {
        // Simplified flash loan that accepts any excess as donation
        let required_amount = amount + (amount * 9 / 10000); // 0.09% fee

        assert(repay_amount >= required_amount, 'Underpaid');

        // VULNERABILITY: Uncapped donation accepted
        let excess = repay_amount - required_amount;
        if excess > 0 {
            // Update reserve with full donation
            let new_reserve = self.reserve_balance.read() + excess;
            self.reserve_balance.write(new_reserve);

            // Update accumulator with no bounds checking
            self._update_accumulator();
        }
    }

    // VULNERABILITY: Special empty market case allows any first deposit amount
    fn _calculate_shares(self: @ContractState, amount: u256) -> u256 {
        let total_supply = self.total_supply.read();

        // If no shares exist yet, use the amount directly
        if total_supply == 0 {
            return amount;
        }

        // Calculate shares based on current accumulator
        let accumulator = self.lending_accumulator.read();
        (amount * SCALE) / accumulator
    }

    fn _update_accumulator(ref self: ContractState) {
        let total_supply = self.total_supply.read();

        // Skip if no supply
        if total_supply == 0 {
            return;
        }

        let reserve = self.reserve_balance.read();

        // VULNERABILITY: No bounds checking on accumulator
        let new_accumulator = (reserve * SCALE) / total_supply;
        self.lending_accumulator.write(new_accumulator);
    }
}
