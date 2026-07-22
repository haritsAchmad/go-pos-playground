import Swal from 'sweetalert2'
import type { Ref } from 'vue'

type Submit=(action:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>,confirmation?:string|false)=>Promise<boolean>

export function useItems(options:{api:any,data:any,filters:any,editing:any,modal:Ref<null|'item'|'customer'|'supplier'>,submit:Submit,findOrCreateMaster:(table:'categories'|'brands'|'units',values:any[],name:string)=>Promise<any>}){
  const {api,data,filters,editing,modal,submit,findOrCreateMaster}=options
  const itemForm=reactive<any>({sku:'',name:'',description:'',supplier_id:null,category_id:null,brand_id:null,brand_name:'',unit_id:null,stock:0,cost:0,price:0})
  const itemImport=ref<HTMLInputElement|null>(null)
  const filteredItems=computed(()=>data.items.filter((v:any)=>{
    const query=filters.itemSearch.toLowerCase();const matchesQuery=!query||[v.sku,v.name,v.category_name,v.brand_name].some(x=>String(x||'').toLowerCase().includes(query));
    const matchesCategory=!filters.category||String(v.category_id)===filters.category;
    const matchesStock=!filters.stock||(filters.stock==='empty'?Number(v.stock)===0:filters.stock==='low'?Number(v.stock)>0&&Number(v.stock)<=5:Number(v.stock)>5);
    return matchesQuery&&matchesCategory&&matchesStock
  }))

  async function loadItems(){data.items=await api.items()}
  function nullableNumber(v:any){return v?Number(v):null}
  async function saveItem(){const saved=await submit(async()=>{const brandID=await findOrCreateMaster('brands',data.brands,itemForm.brand_name);const payload={...itemForm,supplier_id:nullableNumber(itemForm.supplier_id),category_id:nullableNumber(itemForm.category_id),brand_id:brandID,unit_id:nullableNumber(itemForm.unit_id),stock:Number(itemForm.stock),cost:Number(itemForm.cost),price:Number(itemForm.price)};return editing.item?api.updateItem(editing.item,payload):api.createItem(payload)},editing.item?'Barang berhasil diubah':'Barang berhasil ditambahkan',[loadItems]);if(saved)cancelItem()}
  function openItem(v:any=null){cancelItem(false);if(v){editing.item=v.id;Object.assign(itemForm,v)}modal.value='item'}
  function editItem(v:any){openItem(v)}
  function cancelItem(close=true){editing.item=null;Object.assign(itemForm,{sku:'',name:'',description:'',supplier_id:null,category_id:null,brand_id:null,brand_name:'',unit_id:null,stock:0,cost:0,price:0});if(close)modal.value=null}
  async function removeItem(v:any){const result=await Swal.fire({icon:'warning',title:`Hapus ${v.name}?`,text:'Data yang sudah dihapus tidak dapat dikembalikan.',showCancelButton:true,confirmButtonText:'Hapus',cancelButtonText:'Batal',confirmButtonColor:'#b8322a'});if(result.isConfirmed)await submit(()=>api.deleteItem(v.id),'Barang berhasil dihapus',[loadItems])}

  return {itemForm,itemImport,filteredItems,loadItems,saveItem,openItem,editItem,cancelItem,removeItem}
}
