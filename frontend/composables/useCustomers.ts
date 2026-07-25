import Swal from 'sweetalert2'
import type { Ref } from 'vue'

type Submit=(action:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>,confirmation?:string|false)=>Promise<boolean>

export function useCustomers(options:{api:any,data:any,transactionForm:any,filters:any,editing:any,modal:Ref<null|'item'|'customer'|'supplier'>,submit:Submit}){
  const {api,data,transactionForm,filters,editing,modal,submit}=options
  const customerForm=reactive<any>({code:'',name:'',phone:'',address:''})
  const customerImport=ref<HTMLInputElement|null>(null)
  const filteredCustomers=computed(()=>data.customers.filter((v:any)=>!filters.customerSearch||[v.code,v.name,v.phone].some(x=>String(x||'').toLowerCase().includes(filters.customerSearch.toLowerCase()))))

  async function loadCustomers(){data.customers=await api.customers();if(!transactionForm.customer_id)transactionForm.customer_id=data.customers.find((c:any)=>c.code==='UMUM')?.id}
  async function saveCustomer(){const id=editing.customer;if(await submit(()=>id?api.updateCustomer(id,customerForm):api.createCustomer(customerForm),`Pelanggan berhasil ${id?'diubah':'ditambahkan'}`,[loadCustomers]))closeCustomer()}
  function openCustomer(v:any=null){if(v?.code==='UMUM')return;editing.customer=v?.id||null;Object.keys(customerForm).forEach(k=>customerForm[k]='');if(v)Object.assign(customerForm,v);modal.value='customer'}
  function editCustomer(v:any){openCustomer(v)}
  function closeCustomer(){editing.customer=null;Object.keys(customerForm).forEach(k=>customerForm[k]='');modal.value=null}
  async function removeCustomer(v:any){if(v.code==='UMUM')return;const result=await Swal.fire({icon:'warning',title:`Hapus ${v.name}?`,showCancelButton:true,confirmButtonText:'Hapus',cancelButtonText:'Batal',confirmButtonColor:'#b8322a'});if(result.isConfirmed)await submit(()=>api.deleteCustomer(v.id),'Pelanggan berhasil dihapus',[loadCustomers])}

  watch(()=>customerForm.phone,v=>{const clean=String(v||'').replace(/\D/g,'');if(v!==clean)customerForm.phone=clean})

  return {customerForm,customerImport,filteredCustomers,loadCustomers,saveCustomer,openCustomer,editCustomer,closeCustomer,removeCustomer}
}
