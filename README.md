This repository is a Liquidity Module for the XRPL DEX.

Liquidity Modules allow relayers to instantly fill their orderbooks and bootstrap liquidity with minimal effort. This particular module uses a [Constant Product Market Making Model](https://github.com/yxxyun/amm-bots#constant-product-amm), discussed in greater detail below.

#### Using this repository

This repository is designed to start a market making bot to provide liquidity on XRPL DEX.

***

# AMM Bots

Automated Market Making (AMM) bots provide liquidity to a marketplace through use of algorithmic market making.

## Running the AMM Bot

This repository is setup to run naturally on the XRPL DEX. For use on other marketplaces, you could modify the code.

## Algorithm Variables

### Constant Product AMM

This bot runs a "constant product market maker model" (popularized in the DeFi community by Uniswap). In short, this model generates a full orderbook based on an initial price for the market. Every transaction that occurs on this market will adjust the prices of the market accordingly. It's a basic supply and demand automated market making system.

- Buying large amounts of the base token will increase the price
- Selling large amounts of the base token will decrease the price

A typical Constant Market Making model has a continuous price curve. This bot discretizes the continuous price curve and creates a number of limit orders to simulate the curve. The order price is limited between `maxPrice` and `minPrice`. The price difference between adjacent orders is `priceGap`.

![Image](assets/const_product_graph.png)
([Image Source](https://medium.com/scalar-capital/uniswap-a-unique-exchange-f4ef44f807bf))


Constant product algorithms have a disadvantage of low inventory utilization. For example, by default it only uses 5% of your inventory when the price increases 10%. `expandInventory` can help you add depth near the current price.

 - `maxPrice` Max order price
 - `minPrice` Min order price
 - `priceGap` Price difference rate between adjacent orders. For example, ask price increases by 2% and bid price decreases by 2% if `priceGap=0.02`.
 - `expandInventory` Multiply your order sizes linearly. For example, all order sizes will be tripled if `expandInventory=3`.

#### Liquidity Sourcing In Constant Product AMM

Constant Product AMM requires a single source of funds for both the base and quote token. As such, the general flow of setting up the AMM bot from scratch is:

- Prepare an address that holds both the BASE and QUOTE token of the trading pair you want to provide liquidity for
- Set the initial parameters for the bot
  - This determines the initial price, spread, sensitivity, etc.
- Run the bot

Upon running the bot, it will generate an orderbook for your marketplace. The orderbook will appear to be static, but every trade will shift the market accordingly.

#### Further Information On Constant Product AMM

For more information on these models, Scalar Capital provided a [detailed analysis of constant market making](https://medium.com/scalar-capital/uniswap-a-unique-exchange-f4ef44f807bf).

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details

