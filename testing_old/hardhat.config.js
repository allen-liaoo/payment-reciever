require("@nomicfoundation/hardhat-ignition-ethers");

module.exports = {
    networks: {
        hardhat: {
            chainId: 1234,
            mining: {
              auto: true
            }
        },
    },
    solidity: {
        version: "0.4.17" // See ~/.solc-select/artifacts/ ; 0.8.25
    }
}