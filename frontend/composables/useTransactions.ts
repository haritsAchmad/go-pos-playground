import Swal from 'sweetalert2'
import type { Ref } from 'vue'

type Submit=(action:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>,confirmation?:string|false)=>Promise<boolean>

export function useTransactions(options:{api:any,data:any,active:Ref<string>,editing:any,filters:any,submit:Submit,reloadDashboard:()=>Promise<void>,reloadItems:()=>Promise<void>,reloadDebts:()=>Promise<void>,money:(value:number)=>string,printDocument:(mode:'monthly'|'receipt')=>Promise<void>}){
  const {api,data,active,editing,filters,submit,reloadDashboard,reloadItems,reloadDebts,money,printDocument}=options
  const transactionForm=reactive<any>({customer_id:null,supplier_id:null,payment_method_id:null,paid_amount:0,notes:'',items:[{item_id:null,quantity:1,unit_price:0}]})
  const transactionContext=ref<'sale'|'purchase'>('sale')
  const expandedTransaction=ref<number|null>(null)
  const selectedReceipt=ref<any>(null)
  const editDraft=useState<any|null>('transaction-edit-draft',()=>null)
  const lineTotal=computed(()=>transactionForm.items.reduce((sum:number,line:any)=>sum+(Number(line.quantity)||0)*(Number(line.unit_price)||0),0))
  const changeAmount=computed(()=>Number(transactionForm.paid_amount||0)-lineTotal.value)
  const transactionItems=computed(()=>active.value==='purchase'?data.items.filter((v:any)=>Number(v.supplier_id)===Number(transactionForm.supplier_id)):data.items)
  const filteredTransactions=computed(()=>data.transactions.filter((v:any)=>{const query=filters.historySearch.toLowerCase();const matchesSearch=!query||[v.invoice_no,v.customer_name,v.supplier_name,v.notes].some(x=>String(x||'').toLowerCase().includes(query));const matchesType=!filters.historyType||v.transaction_type===filters.historyType;const matchesStatus=!filters.historyStatus||(filters.historyStatus==='ACTIVE'?v.status==='ACTIVE':filters.historyStatus==='VOID'?v.status==='VOID':v.payment_status===filters.historyStatus);return matchesSearch&&matchesType&&matchesStatus}))

  async function loadTransactions(){data.transactions=await api.transactions()}
  async function voidTransaction(v:any){const stockEffect=v.transaction_type==='SALE'?'Stok barang penjualan akan dikembalikan.':'Stok dari pembelian akan dikurangi; pembatalan ditolak jika stok sudah terpakai.';const result=await Swal.fire({icon:'warning',title:`Batalkan ${v.invoice_no}?`,text:stockEffect,input:'textarea',inputLabel:'Alasan pembatalan',inputPlaceholder:'Tulis alasan (minimal 5 karakter)',showCancelButton:true,confirmButtonText:'Ya, batalkan',cancelButtonText:'Kembali',confirmButtonColor:'#b8322a',inputValidator:(value)=>!value||value.trim().length<5?'Alasan pembatalan minimal 5 karakter.':undefined});if(!result.isConfirmed)return;await submit(()=>api.voidTransaction(v.id,result.value.trim()),`Transaksi dibatalkan. ${stockEffect}`,[reloadDashboard,reloadItems,loadTransactions,reloadDebts],false)}
  function addLine(){transactionForm.items.push({item_id:null,quantity:1,unit_price:0})}
  function resetTransaction(context:string=active.value){editing.transaction=null;transactionContext.value=context==='purchase'?'purchase':'sale';Object.assign(transactionForm,{customer_id:context==='sale'?data.customers.find((c:any)=>c.code==='UMUM')?.id||null:null,supplier_id:null,payment_method_id:null,paid_amount:0,notes:'',items:[{item_id:null,quantity:1,unit_price:0}]})}
  async function editTransaction(v:any){if(v.status!=='ACTIVE')return;const context=v.transaction_type==='SALE'?'sale':'purchase';editDraft.value={id:v.id,context,form:{customer_id:v.customer_id,supplier_id:v.supplier_id,payment_method_id:v.payment_method_id,paid_amount:v.amount_received,notes:v.notes||'',items:v.items.map((line:any)=>({item_id:line.item_id,quantity:line.quantity,unit_price:line.unit_price}))}};await navigateTo(context==='sale'?'/kasir':'/pembelian')}
  async function saveTransaction(type:'SALE'|'PURCHASE'){const payload={...transactionForm,transaction_type:type,customer_id:type==='SALE'?Number(transactionForm.customer_id):null,supplier_id:type==='PURCHASE'?Number(transactionForm.supplier_id):null,payment_method_id:transactionForm.payment_method_id?Number(transactionForm.payment_method_id):null,paid_amount:Number(transactionForm.paid_amount),items:transactionForm.items.map((v:any)=>({...v,item_id:Number(v.item_id),quantity:Number(v.quantity),unit_price:Number(v.unit_price)}))};const id=editing.transaction;if(await submit(()=>id?api.updateTransaction(id,payload):api.createTransaction(payload),`${type==='SALE'?'Penjualan':'Pembelian'} berhasil ${id?'diubah':'dicatat'}`,[reloadDashboard,reloadItems,loadTransactions,reloadDebts],`${id?'Simpan perubahan':'Simpan'} ${type==='SALE'?'penjualan':'pembelian'} sebesar ${money(lineTotal.value)}?`))resetTransaction()}
  function chooseItem(line:any){const item=data.items.find((v:any)=>v.id===Number(line.item_id));if(item)line.unit_price=active.value==='purchase'?item.cost:item.price}
  function printReceipt(transaction:any){selectedReceipt.value=transaction;printDocument('receipt')}

  if(editDraft.value&&editDraft.value.context===active.value){editing.transaction=editDraft.value.id;transactionContext.value=editDraft.value.context;Object.assign(transactionForm,editDraft.value.form);editDraft.value=null}

  watch(()=>transactionForm.supplier_id,()=>{if(active.value==='purchase'){transactionForm.items=transactionForm.items.map((line:any)=>transactionItems.value.some((v:any)=>v.id===Number(line.item_id))?line:{item_id:null,quantity:1,unit_price:0})}})

  return {transactionForm,transactionContext,expandedTransaction,selectedReceipt,lineTotal,changeAmount,transactionItems,filteredTransactions,loadTransactions,voidTransaction,addLine,resetTransaction,editTransaction,saveTransaction,chooseItem,printReceipt}
}
