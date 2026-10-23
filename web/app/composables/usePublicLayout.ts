// One viewport rule, shared by the shell and page views. Data controllers live
// above either presentation, so switching views never creates another query.
export function usePublicLayout() {
 const mobile=useState('ytv-mobile-layout',()=>false);
 const ready=useState('ytv-layout-ready',()=>false);
 let query:MediaQueryList|undefined;
 const update=()=>{mobile.value=query?.matches||false;ready.value=true;};
 onMounted(()=>{query=window.matchMedia('(max-width: 767px)');update();query.addEventListener('change',update);});
 onBeforeUnmount(()=>query?.removeEventListener('change',update));
 return {mobile,ready};
}
