#[starknet::contract]
mod MinimalSecureMarket {
    use starknet::{ContractAddress, get_caller_address, get_block_timestamp};
    use starknet::contract_address::ContractAddressZeroable;

    // Storage variables
    #[storage]
    struct Storage {
        reserve_balance: u256,
        total_supply: u256,
        lending_accumulator: u256,
        user_balances: LegacyMap<ContractAddress, u256>,
        
        // SAFEGUARD: Market activation flag and admin
        is_active: bool,
        admin: ContractAddress,
        min_deposit: u256,
        
        // SAFEGUARD: Race condition protections
        reentrancy_guard: bool,
        activation_timestamp: u64,
        grace_period: u64,
    }

    const SCALE: u256 = 1000000000000000000; // 1e18
    const MIN_ACCUMULATOR: u256 = 900000000000000000; // 0.9
    const MAX_ACCUMULATOR: u256 = 2000000000000000000; // 2.0

    #[constructor]
    fn constructor(ref self: ContractState, admin: ContractAddress) {
        self.lending_accumulator.write(SCALE); // Initialize to 1.0
        self.is_active.write(false); // Start inactive
        self.admin.write(admin);
        self.min_deposit.write(SCALE / 10); // 0.1 tokens minimum
        self.reentrancy_guard.write(false);
        self.grace_period.write(3600); // 1 hour grace period after activation
    }

    // SAFEGUARD: Explicit market activation with minimum liquidity and reentrancy protection
    #[external(v0)]
    fn activate_market(ref self: ContractState, initial_deposit: u256) {
        // Check reentrancy guard
        assert(!self.reentrancy_guard.read(), 'Reentrancy detected');
        self.reentrancy_guard.write(true);
        
        let caller = get_caller_address();
        assert(caller == self.admin.read(), 'Not admin');
        assert(!self.is_active.read(), 'Already active');
        assert(initial_deposit >= self.min_deposit.read(), 'Deposit too small');
        
        // Set initial state
        self.reserve_balance.write(initial_deposit);
        self.user_balances.write(caller, initial_deposit);
        self.total_supply.write(initial_deposit);
        
        // Set activation timestamp and active flag
        self.activation_timestamp.write(get_block_timestamp());
        self.is_active.write(true);
        
        // Release reentrancy guard
        self.reentrancy_guard.write(false);
    }

    // SAFEGUARD: Check activation, grace period, and minimum deposit with reentrancy protection
    #[external(v0)]
    fn deposit(ref self: ContractState, amount: u256) -> u256 {
        // Check reentrancy guard
        assert(!self.reentrancy_guard.read(), 'Reentrancy detected');
        self.reentrancy_guard.write(true);
        
        // Check market is active
        assert(self.is_active.read(), 'Market not active');
        
        let caller = get_caller_address();
        let admin = self.admin.read();
        
        // During grace period, only admin can deposit
        if get_block_timestamp() < self.activation_timestamp.read() + self.grace_period.read() {
            assert(caller == admin, 'In grace period');
        }
        
        // Check minimum deposit
        assert(amount >= self.min_deposit.read(), 'Deposit too small');
        
        // Update reserve balance
        let new_reserve = self.reserve_balance.read() + amount;
        self.reserve_balance.write(new_reserve);
        
        // Calculate shares to mint (no special empty case needed)
        let shares = self._calculate_shares(amount);
        
        // Update user's balance and total supply
        let new_balance = self.user_balances.read(caller) + shares;
        self.user_balances.write(caller, new_balance);
        
        let new_supply = self.total_supply.read() + shares;
        self.total_supply.write(new_supply);
        
        // Release reentrancy guard
        self.reentrancy_guard.write(false);
        
        shares
    }

    #[external(v0)]
    fn flash_loan(ref self: ContractState, amount: u256, repay_amount: u256) {
        // Check reentrancy guard
        assert(!self.reentrancy_guard.read(), 'Reentrancy detected');
        self.reentrancy_guard.write(true);
        
        // Check market is active
        assert(self.is_active.read(), 'Market not active');
        
        // Check grace period is over
        assert(get_block_timestamp() >= self.activation_timestamp.read() + self.grace_period.read(), 
               'In grace period');
        
        let required_amount = amount + (amount * 9 / 10000); // 0.09% fee
        assert(repay_amount >= required_amount, 'Underpaid');
        
        // SAFEGUARD: Cap donation amount
        let excess = repay_amount - required_amount;
        if excess > 0 {
            let reserve = self.reserve_balance.read();
            
            // Cap donation to 0.1% of reserve
            let max_donation = reserve / 1000;
            let accepted_donation = if excess > max_donation { max_donation } else { excess };
            
            // Update reserve with capped donation
            self.reserve_balance.write(reserve + accepted_donation);
            
            // Update accumulator with bounds checking
            self._update_accumulator_safe();
        }
        
        // Release reentrancy guard
        self.reentrancy_guard.write(false);
    }

    fn _calculate_shares(self: @ContractState, amount: u256) -> u256 {
        // No special empty market case - market activation ensures non-zero supply
        let accumulator = self.lending_accumulator.read();
        (amount * SCALE) / accumulator
    }

    // SAFEGUARD: Bounds checking on accumulator
    fn _update_accumulator_safe(ref self: ContractState) {
        let total_supply = self.total_supply.read();
        let reserve = self.reserve_balance.read();
        
        // Calculate new accumulator value
        let calculated = (reserve * SCALE) / total_supply;
        
        // Apply bounds check
        let new_accumulator = if calculated < MIN_ACCUMULATOR {
            MIN_ACCUMULATOR
        } else if calculated > MAX_ACCUMULATOR {
            MAX_ACCUMULATOR
        } else {
            calculated
        };
        
        self.lending_accumulator.write(new_accumulator);
    }
}
