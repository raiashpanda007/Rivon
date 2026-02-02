"use client"
import ApiCaller, { RequestType } from "@workspace/api-caller";
import { Wallet } from "lucide-react"
import { useEffect, useState } from "react";
function WalletIcon() {
  const [walletAmount, setWalletAmount] = useState(0);
  async function GetWalletStatus() {
    const response = await ApiCaller<undefined, { id: string, userId: string, balance: number }>({
      requestType: RequestType.GET,
      paths: ["api", "rivon", "wallet", "me"]
    });
    if (response.ok) {
      setWalletAmount(response.response.data.balance);
    }
  }
  useEffect(() => {
    const run = async () => {
      await GetWalletStatus();
    }
    run();
  }, []);
  return (
    <div className="h-11/12 opacity-85 border  flex items-center space-x-1">
      <Wallet className="text-orange-500" />
      <span className="font-semibold">
        $ {(walletAmount / 100).toFixed(2)}

      </span>
    </div>
  )
}



export default WalletIcon;
