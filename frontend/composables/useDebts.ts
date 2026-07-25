type Submit=(action:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>,confirmation?:string|false)=>Promise<boolean>

export function useDebts(options:{api:any,data:any,submit:Submit,reloadDashboard:()=>Promise<void>,reloadTransactions:()=>Promise<void>}){
  const {api,data,submit,reloadDashboard,reloadTransactions}=options
  const debtPayments=reactive<Record<number,number>>({})

  async function loadDebts(){data.debts=await api.debts()}
  async function payDebt(v:any){const amount=Number(debtPayments[v.id]||0);if(await submit(()=>api.payDebt(v.id,{amount}),'Pembayaran piutang berhasil dicatat',[reloadDashboard,reloadTransactions,loadDebts]))debtPayments[v.id]=0}

  return {debtPayments,loadDebts,payDebt}
}
