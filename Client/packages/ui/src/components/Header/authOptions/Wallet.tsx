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
    GetWalletStatus();
    const id = setInterval(GetWalletStatus, 30_000);
    return () => clearInterval(id);
  }, []);
  return (
    <a href="/wallet" className="h-11/12 opacity-85 border flex items-center space-x-1 hover:opacity-100 transition-opacity">
      <Wallet className="text-orange-500" />
      <span className="font-semibold">
        $ {(walletAmount / 100).toFixed(2)}
      </span>
    </a>
  )
}



export default WalletIcon;
