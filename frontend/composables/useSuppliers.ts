import Swal from 'sweetalert2'
import type { Ref } from 'vue'

type Submit=(action:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>,confirmation?:string|false)=>Promise<boolean>

export function useSuppliers(options:{api:any,data:any,filters:any,editing:any,modal:Ref<null|'item'|'customer'|'supplier'>,submit:Submit}){
  const {api,data,filters,editing,modal,submit}=options
  const supplierForm=reactive<any>({code:'',name:'',phone:'',address:''})
  const supplierImport=ref<HTMLInputElement|null>(null)
  const filteredSuppliers=computed(()=>data.suppliers.filter((v:any)=>!filters.supplierSearch||[v.code,v.name,v.phone,v.address].some(x=>String(x||'').toLowerCase().includes(filters.supplierSearch.toLowerCase()))))

  async function loadSuppliers(){data.suppliers=await api.suppliers()}
  async function saveSupplier(){const id=editing.supplier;if(await submit(()=>id?api.updateSupplier(id,supplierForm):api.createSupplier(supplierForm),`Supplier berhasil ${id?'diubah':'ditambahkan'}`,[loadSuppliers]))closeSupplier()}
  function openSupplier(v:any=null){editing.supplier=v?.id||null;Object.keys(supplierForm).forEach(k=>supplierForm[k]='');if(v)Object.assign(supplierForm,v);modal.value='supplier'}
  function editSupplier(v:any){openSupplier(v)}
  function closeSupplier(){editing.supplier=null;Object.keys(supplierForm).forEach(k=>supplierForm[k]='');modal.value=null}
  async function removeSupplier(v:any){const result=await Swal.fire({icon:'warning',title:`Hapus ${v.name}?`,showCancelButton:true,confirmButtonText:'Hapus',cancelButtonText:'Batal',confirmButtonColor:'#b8322a'});if(result.isConfirmed)await submit(()=>api.deleteSupplier(v.id),'Supplier berhasil dihapus',[loadSuppliers])}

  watch(()=>supplierForm.phone,v=>{const clean=String(v||'').replace(/\D/g,'');if(v!==clean)supplierForm.phone=clean})

  return {supplierForm,supplierImport,filteredSuppliers,loadSuppliers,saveSupplier,openSupplier,editSupplier,closeSupplier,removeSupplier}
}
