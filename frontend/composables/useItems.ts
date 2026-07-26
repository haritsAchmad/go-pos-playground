import Swal from 'sweetalert2'
import type { Ref } from 'vue'

type Submit=(action:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>,confirmation?:string|false)=>Promise<boolean>
type SoftDelete=(remove:()=>Promise<any>,restore:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>)=>Promise<boolean>

export function useItems(options:{api:any,data:any,filters:any,editing:any,modal:Ref<null|'item'|'customer'|'supplier'>,submit:Submit,softDelete:SoftDelete,findOrCreateMaster:(table:'categories'|'brands'|'units',values:any[],name:string)=>Promise<any>}){
  const {api,data,filters,editing,modal,submit,softDelete,findOrCreateMaster}=options
  const emptyItem=()=>({sku:'',name:'',description:'',supplier_id:null,category_id:null,brand_id:null,brand_name:'',unit_id:null,base_unit_id:null,units_per_package:1,allow_retail:false,stock:0,cost:0,price:0,retail_cost:0,retail_price:0})
  const itemForm=reactive<any>(emptyItem())
  const itemImport=ref<HTMLInputElement|null>(null)
  const filteredItems=computed(()=>data.items.filter((v:any)=>{
    const query=filters.itemSearch.toLowerCase();const matchesQuery=!query||[v.sku,v.name,v.category_name,v.brand_name].some(x=>String(x||'').toLowerCase().includes(query));
    const matchesCategory=!filters.category||String(v.category_id)===filters.category;
    const matchesStock=!filters.stock||(filters.stock==='empty'?Number(v.stock)===0:filters.stock==='low'?Number(v.stock)>0&&Number(v.stock)<=5:Number(v.stock)>5);
    return matchesQuery&&matchesCategory&&matchesStock
  }))

  function stockDisplay(item:any){const factor=Math.max(1,Number(item.units_per_package)||1);if(factor===1)return`${item.stock} ${item.unit_name||''}`.trim();const packages=Math.floor(Number(item.stock)/factor),remainder=Number(item.stock)%factor;return`${packages} ${item.unit_name||'kemasan'}${remainder?` + ${remainder} ${item.base_unit_name||'satuan'}`:''}`}
  async function loadItems(){data.items=(await api.items()).map((item:any)=>({...item,stock_display:stockDisplay(item)}))}
  function nullableNumber(v:any){return v?Number(v):null}
  async function saveItem(){const saved=await submit(async()=>{const brandID=await findOrCreateMaster('brands',data.brands,itemForm.brand_name);const factor=Math.max(1,Number(itemForm.units_per_package)||1);const payload={...itemForm,supplier_id:nullableNumber(itemForm.supplier_id),category_id:nullableNumber(itemForm.category_id),brand_id:brandID,unit_id:nullableNumber(itemForm.unit_id),base_unit_id:nullableNumber(itemForm.base_unit_id||itemForm.unit_id),units_per_package:factor,allow_retail:factor>1&&Boolean(itemForm.allow_retail),stock:Number(itemForm.stock),cost:Number(itemForm.cost),price:Number(itemForm.price),retail_cost:factor===1?Number(itemForm.cost):Number(itemForm.retail_cost),retail_price:factor===1?Number(itemForm.price):Number(itemForm.retail_price)};return editing.item?api.updateItem(editing.item,payload):api.createItem(payload)},editing.item?'Barang berhasil diubah':'Barang berhasil ditambahkan',[loadItems]);if(saved)cancelItem()}
  function openItem(v:any=null){cancelItem(false);if(v){editing.item=v.id;Object.assign(itemForm,v)}modal.value='item'}
  function editItem(v:any){openItem(v)}
  function cancelItem(close=true){editing.item=null;Object.assign(itemForm,emptyItem());if(close)modal.value=null}
  async function removeItem(v:any){const result=await Swal.fire({icon:'warning',title:`Hapus ${v.name}?`,text:'Data dapat dipulihkan kembali sesaat setelah dihapus.',showCancelButton:true,confirmButtonText:'Hapus',cancelButtonText:'Batal',confirmButtonColor:'#b8322a'});if(result.isConfirmed)await softDelete(()=>api.deleteItem(v.id),()=>api.restoreItem(v.id),`Barang ${v.name} dihapus`,[loadItems])}
  async function restoreDeletedItem(){const rows=await api.deletedItems();if(!rows.length){await Swal.fire({icon:'info',title:'Tidak ada barang terhapus'});return}const result=await Swal.fire({title:'Pulihkan barang',width:'min(720px, calc(100vw - 32px))',padding:'2rem',customClass:{popup:'restore-dialog',input:'restore-dialog-input'},input:'select',inputOptions:Object.fromEntries(rows.map((v:any)=>[v.id,`${v.sku} · ${v.name}`])),inputPlaceholder:'Pilih barang',showCancelButton:true,confirmButtonText:'Pulihkan',cancelButtonText:'Batal',inputValidator:(v)=>v?undefined:'Pilih barang terlebih dahulu'});if(result.isConfirmed)await submit(()=>api.restoreItem(Number(result.value)),'Barang berhasil dipulihkan',[loadItems])}

  return {itemForm,itemImport,filteredItems,loadItems,saveItem,openItem,editItem,cancelItem,removeItem,restoreDeletedItem}
}
