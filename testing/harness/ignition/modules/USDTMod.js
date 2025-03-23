const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");
const { Wallet } = require('ethers');
const chain_info = require('../../../chain_info.json');

module.exports = buildModule("USDTMod", (m) => {
    const conCreator = new Wallet(chain_info["USDTCreatorPK"])

    // deploy USDT's contract at conCreator address
    const amount = 100000000000
    const usdt = m.contract("TetherToken", [amount, "Tether USD", "USD", 6], { 
        from: conCreator.address
      });
    console.log("deploying contract:", usdt)

    return { usdt };
});
