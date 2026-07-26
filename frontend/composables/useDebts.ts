import Swal from 'sweetalert2'

type Submit=(action:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>,confirmation?:string|false)=>Promise<boolean>

export function useDebts(options:{api:any,data:any,submit:Submit,reloadDashboard:()=>Promise<void>,reloadTransactions:()=>Promise<void>}){
  const {api,data,submit,reloadDashboard,reloadTransactions}=options
  const debtPayments=reactive<Record<number,number>>({})
  const paymentHistories=reactive<Record<number,any[]>>({})
  const expandedDebt=ref<number|null>(null)

  async function loadDebts(){data.debts=await api.debts()}
  async function payDebt(v:any){const amount=Number(debtPayments[v.id]||0);if(await submit(()=>api.payDebt(v.id,{amount}),'Pembayaran piutang berhasil dicatat',[reloadDashboard,reloadTransactions,loadDebts]))debtPayments[v.id]=0}
  async function togglePaymentHistory(v:any){if(expandedDebt.value===v.id){expandedDebt.value=null;return}paymentHistories[v.id]=await api.debtPayments(v.id);expandedDebt.value=v.id}
  async function reversePayment(debt:any,payment:any){const result=await Swal.fire({icon:'warning',title:'Batalkan pembayaran?',text:`Pembayaran Rp ${Number(payment.amount).toLocaleString('id-ID')} akan dikembalikan menjadi piutang.`,input:'textarea',inputLabel:'Alasan koreksi',inputPlaceholder:'Contoh: nominal pembayaran salah input',inputAttributes:{maxlength:'500'},showCancelButton:true,confirmButtonText:'Batalkan pembayaran',cancelButtonText:'Kembali',confirmButtonColor:'#b8322a',inputValidator:value=>String(value||'').trim().length>=5?undefined:'Alasan minimal 5 karakter'});if(!result.isConfirmed)return;const done=await submit(()=>api.reverseDebtPayment(debt.id,payment.id,String(result.value).trim()),'Pembayaran berhasil dibatalkan',[reloadDashboard,reloadTransactions,loadDebts]);if(done)paymentHistories[debt.id]=await api.debtPayments(debt.id)}

  return {debtPayments,paymentHistories,expandedDebt,loadDebts,payDebt,togglePaymentHistory,reversePayment}
}
