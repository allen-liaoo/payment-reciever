const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");
const { Wallet } = require('ethers');
const chain_info = require('../../../chain_info.json');

module.exports = buildModule("USDTMod", (m) => {
    const conCreator = new Wallet(chain_info["contractCreatorPK"])

    // deploy contract at conCreator address
    const amount = 100000000000
    const usdt = m.contract("TetherToken", [amount, "Tether USD", "USD", 6], { 
        from: conCreator.address
      });
    console.log("deploying contract:", usdt)

    // const middlewares = []
    // for (const middlewarePK of chain_info.middlewarePKs) {
    //   middlewares.push(new Wallet(middlewarePK))
    // }

    // TODO: Move this to test site
    // for (const middleware of middlewares) {
    //   m.call(usdt, "transferFrom", [conCreator.address, middleware.address, Math.trunc(amount / middlewares.length)])
    // }

    // for (const middleware of middlewares) {
    //   console.log(middleware.address + ": " + m.staticCall(usdt, "balanceOf", [middleware.address]))
    // }

    return { usdt };
});
