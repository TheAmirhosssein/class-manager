(() => {
  'use strict';
  const $=(s,root=document)=>root.querySelector(s),$$=(s,root=document)=>[...root.querySelectorAll(s)];
  const esc=v=>String(v??'').replace(/[&<>"']/g,x=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[x]));
  const fa=v=>String(v).replace(/\d/g,x=>'۰۱۲۳۴۵۶۷۸۹'[x]);
  const days=['شنبه','یکشنبه','دوشنبه','سه‌شنبه','چهارشنبه','پنجشنبه','جمعه'];
  const today=()=>{const d=new Date();return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`;};
  const prettyDate=v=>{try{return new Intl.DateTimeFormat('fa-IR',{dateStyle:'medium'}).format(new Date(v+'T12:00:00'));}catch{return v;}};
  const BRAND='کلاسینو';
  const DEMO_AUTH_KEY='classino.demoAuth';
  let classes=[],schedule=[],school={name:'مدرسه نمونه'},routeVersion=0,currentModalOp=0,toastTimer,user=null;
  const classBy=id=>classes.find(x=>x.id===id);
  const classOptions=()=>classes.map(x=>`<option value="${esc(x.id)}">${esc(x.title)}</option>`).join('');
  const dayOptions=()=>days.map((x,i)=>`<option value="${i}">${x}</option>`).join('');
  const error=e=>`<div class="error-message" role="alert">${esc(e.message||'خطایی رخ داد.')}</div>`;
  function toast(message){$('#toast').textContent=message;$('#toast').hidden=false;clearTimeout(toastTimer);toastTimer=setTimeout(()=>$('#toast').hidden=true,4500);}
  function showModal(html){currentModalOp++;$('#modal-content').innerHTML=html;if(!$('#modal').open)$('#modal').showModal();return currentModalOp;}
  function closeModal(){currentModalOp++;$('#modal').close();}
  $('.modal-close')?.addEventListener('click',closeModal);
  $('#modal')?.addEventListener('click',e=>{if(e.target===$('#modal')){const r=e.target.getBoundingClientRect();if(e.clientX<r.left||e.clientX>r.right||e.clientY<r.top||e.clientY>r.bottom)closeModal();}});
  $('#modal')?.addEventListener('cancel',()=>currentModalOp++);
  function setSchool(){ $('#school-label').textContent=school.name; }
  function isLoggedIn(){return !!user&&['teacher','admin'].includes(user.role);}
  function setView(mode){
    const panel=mode==='panel';
    document.body.classList.toggle('app-panel',panel);
    document.body.classList.toggle('app-auth',!panel);
    $('#login-view').hidden=panel;
    $('#panel-view').hidden=!panel;
  }
  function updateHeaderAuth(){
    const btn=$('#login-button');
    if(!btn)return;
    if(isLoggedIn()){
      btn.textContent=user.name;
      btn.setAttribute('href','#login/account');
    }else{
      btn.innerHTML='ورود <span aria-hidden="true">↗</span>';
      btn.setAttribute('href','#login');
    }
  }
  async function refreshSession(){
    if(APP_CONFIG.demo){
      if(sessionStorage.getItem(DEMO_AUTH_KEY)==='1')user={id:'demo',name:'مدرس نمونه',role:'teacher'};
      else user=null;
      return;
    }
    const session=await API.session();
    user=session?.user||null;
  }
  function completeLogin(nextUser){
    user=nextUser;
    if(APP_CONFIG.demo)sessionStorage.setItem(DEMO_AUTH_KEY,'1');
    updateHeaderAuth();
    const next=sessionStorage.getItem('classino.afterLogin')||'panel';
    sessionStorage.removeItem('classino.afterLogin');
    location.hash=next.startsWith('panel')?next:'panel';
  }
  async function logout(){
    try{
      if(!APP_CONFIG.demo)await API.logout();
    }catch(e){toast(e.message);return;}
    user=null;
    if(APP_CONFIG.demo)sessionStorage.removeItem(DEMO_AUTH_KEY);
    updateHeaderAuth();
    location.hash='login';
  }
  function renderLoginPage(sub){
    setView('login');
    document.title=(sub==='account'?'حساب کاربری':'ورود')+' | '+BRAND;
    const logged=isLoggedIn();
    const account=sub==='account'&&logged;
    $('#login-account').hidden=!account;
    $('#login-forms').hidden=account;
    if(account){
      $('#account-name').textContent=user.name;
      return;
    }
    if(APP_CONFIG.demo){$('#demo-auth-note').hidden=false;$('#demo-enter').hidden=false;}else{$('#demo-auth-note').hidden=true;$('#demo-enter').hidden=true;}
    $('#login-heading').textContent='ورود به کلاسینو';
    $('#auth-status').textContent='';
    $('#auth-form').reset();
  }
  async function submitAuth(e){
    e.preventDefault();
    const btn=$('#auth-submit');
    const status=$('#auth-status');
    btn.disabled=true;
    status.textContent='';
    try{
      if(APP_CONFIG.demo){
        completeLogin({id:'demo',name:'مدرس نمونه',role:'teacher'});
        return;
      }
      const payload=Object.fromEntries(new FormData(e.target));
      const result=await API.login(payload);
      if(!result.user){status.textContent=result.message||'ورود انجام نشد. دوباره تلاش کنید.';return;}
      completeLogin(result.user);
    }catch(err){status.textContent=err.message;}
    finally{btn.disabled=false;}
  }
  async function routePanel(){
    const version=++routeVersion;
    let hash=location.hash.slice(1);
    if(!hash.startsWith('panel'))hash='panel';
    let tab=hash.split('/')[1]||'overview';
    const titles={overview:'نمای کلی',classes:'کلاس‌ها',attendance:'حضور و غیاب',discipline:'انضباط',schedule:'برنامه کلاسی',settings:'تنظیمات مدرسه'};
    if(!titles[tab])tab='overview';
    setView('panel');
    $('#panel-title').textContent=titles[tab];
    document.title=titles[tab]+' | '+BRAND;
    $$('[data-panel]').forEach(x=>x.classList.toggle('active',x.dataset.panel===tab));
    const body=$('#panel-body');body.innerHTML='<p class="help" role="status">در حال دریافت اطلاعات…</p>';
    try {
      await refreshSession();if(version!==routeVersion)return;
      if(!isLoggedIn()){sessionStorage.setItem('classino.afterLogin',hash);location.replace('#login');return;}
      updateHeaderAuth();
      if(tab==='overview')renderOverview();
      if(tab==='classes')renderClasses();
      if(tab==='attendance')await renderAttendance(version);
      if(tab==='discipline')await renderDiscipline(version);
      if(tab==='schedule')await renderSchedule(version);
      if(tab==='settings')renderSettings();
    }catch(e){if(version===routeVersion)body.innerHTML=error(e)+'<div class="quick-links"><button class="btn btn-green" id="panel-retry">تلاش دوباره</button><a href="#login" class="btn btn-outline">ورود به حساب</a></div>';$('#panel-retry')?.addEventListener('click',route);}
  }
  async function route(){
    const hash=location.hash.slice(1)||'login';
    const parts=hash.split('/');
    if(parts[0]==='login'){
      if(parts[1]==='register'){location.replace('#login');return;}
      await refreshSession();
      renderLoginPage(parts[1]==='account'?'account':'login');
      updateHeaderAuth();
      return;
    }
    if(parts[0]==='panel'||hash==='home'){
      if(hash==='home'){location.replace('#login');return;}
      await routePanel();
      return;
    }
    location.replace('#login');
  }
  function renderOverview(){const total=classes.reduce((n,c)=>n+c.studentCount,0);$('#panel-body').innerHTML=`${APP_CONFIG.demo?'<div class="notice">این پنل با اطلاعات نمونه کار می‌کند. تغییرات فقط در همین مرورگر ذخیره می‌شوند.</div>':''}<div class="stats"><div class="stat"><b>${fa(classes.length)}</b><span>کلاس آموزشی</span></div><div class="stat"><b>${fa(total)}</b><span>عضویت در کلاس‌ها</span></div><div class="stat"><b>${fa(schedule.length)}</b><span>بازه در برنامه هفتگی</span></div></div><div class="surface"><div class="eyebrow">سلام، ${esc(user.name)}</div><h2>یک روز تازه، یک فرصت برای یادگیری</h2><p>از اینجا کلاس‌ها را مدیریت کنید، حضور و غیاب را ثبت کنید و برنامه هفته را منظم نگه دارید.</p><div class="quick-links"><a href="#panel/attendance" class="btn btn-green">ثبت حضور و غیاب ←</a><a href="#panel/schedule" class="btn btn-outline">برنامه هفتگی</a></div></div><div class="surface"><h2>مدرسه شما</h2><p>${esc(school.name)}</p><a class="text-link" href="#panel/settings">ویرایش نام مدرسه ←</a></div>`;}
  function renderClasses(){$('#panel-body').innerHTML=`<p class="help">برای ثبت وضعیت آموزشی، کلاس موردنظر را انتخاب کنید.</p><div class="panel-class-grid">${classes.map(c=>`<article class="surface"><span class="badge">پایه ${fa(c.grade)}</span><h3>${esc(c.title)}</h3><p>${esc(c.room)} • ${fa(c.studentCount)} دانش‌آموز</p><button class="btn btn-outline btn-small" data-attend-class="${esc(c.id)}">حضور و غیاب ←</button></article>`).join('')}</div>`;$$('[data-attend-class]').forEach(b=>b.onclick=()=>{sessionStorage.setItem('classino.selectedClass',b.dataset.attendClass);location.hash='panel/attendance';});}
  function selectedClass(){try{const id=sessionStorage.getItem('classino.selectedClass')||sessionStorage.getItem('riazino.selectedClass');return classBy(id)?id:classes[0]?.id||'';}catch{return classes[0]?.id||'';}}
  async function renderAttendance(version){
    $('#panel-body').innerHTML=`<div class="form-row"><label>انتخاب کلاس<select id="attendance-class">${classOptions()}</select></label><label>تاریخ جلسه<input type="date" id="attendance-date" value="${today()}" required></label><button class="btn btn-outline" id="all-present">همه حاضر</button></div><div id="attendance-caption" class="help"></div><div id="attendance-table"></div><div class="save-bar"><button class="btn btn-green" id="attendance-save" disabled>ذخیره حضور و غیاب</button><span id="attendance-status" role="status">وضعیت را برای همه دانش‌آموزان مشخص کنید.</span></div>`;
    $('#attendance-class').value=selectedClass();let loadId=0,roster=[],savedClass='',savedDate='',dirty=false;
    async function load(){const op=++loadId,id=$('#attendance-class').value,date=$('#attendance-date').value;$('#attendance-save').disabled=true;$('#attendance-table').innerHTML='<p class="help">در حال دریافت فهرست…</p>';$('#attendance-status').textContent='';if(!id||!date){$('#attendance-table').innerHTML='<p class="empty">کلاس و تاریخ معتبر انتخاب کنید.</p>';return;}try{const [students,existing]=await Promise.all([API.students(id),API.attendance(id,date)]);if(version!==routeVersion||op!==loadId)return;roster=students;savedClass=id;savedDate=date;dirty=false;$('#attendance-caption').textContent='تاریخ جلسه: '+prettyDate(date);$('#attendance-table').innerHTML=students.length?`<div class="table-wrap"><table><thead><tr><th>ردیف</th><th>نام دانش‌آموز</th><th>وضعیت حضور</th></tr></thead><tbody>${students.map((s,i)=>{const status=existing.records.find(x=>x.studentId===s.id)?.status||'';return `<tr><td>${fa(i+1)}</td><td>${esc(s.name)}</td><td><select data-student="${esc(s.id)}" aria-label="وضعیت ${esc(s.name)}"><option value="">انتخاب وضعیت</option>${[['present','حاضر'],['absent','غایب'],['late','تأخیر'],['excused','غیبت موجه']].map(([v,l])=>`<option value="${v}" ${v===status?'selected':''}>${l}</option>`).join('')}</select></td></tr>`;}).join('')}</tbody></table></div>`:'<p class="empty">دانش‌آموزی در این کلاس ثبت نشده است.</p>';$('#attendance-save').disabled=!students.length;$('#attendance-status').textContent=existing.records.length?'اطلاعات ثبت‌شده بارگذاری شد.':'هنوز برای این جلسه حضوری ثبت نشده است.';}catch(e){if(version===routeVersion&&op===loadId)$('#attendance-table').innerHTML=error(e);}}
    $('#attendance-table').onchange=()=>{dirty=true;$('#attendance-status').textContent='تغییرات هنوز ذخیره نشده‌اند.';};
    let previousClass=$('#attendance-class').value,previousDate=$('#attendance-date').value;
    const change=()=>{if(dirty&&!confirm('تغییرات ذخیره نشده‌اند. از تغییر کلاس یا تاریخ مطمئن هستید؟')){$('#attendance-class').value=previousClass;$('#attendance-date').value=previousDate;return;}previousClass=$('#attendance-class').value;previousDate=$('#attendance-date').value;load();};
    $('#attendance-class').onchange=change;$('#attendance-date').onchange=change;
    $('#all-present').onclick=()=>{$$('[data-student]').forEach(x=>x.value='present');dirty=true;$('#attendance-status').textContent='همه حاضر انتخاب شدند؛ برای ثبت، ذخیره را بزنید.';};
    $('#attendance-save').onclick=async()=>{const records=$$('[data-student]').map(x=>({studentId:x.dataset.student,status:x.value}));if(records.length!==roster.length||records.some(x=>!x.status)){toast('وضعیت همه دانش‌آموزان را مشخص کنید.');return;}const btn=$('#attendance-save');btn.disabled=true;$('#attendance-class').disabled=true;$('#attendance-date').disabled=true;$('#all-present').disabled=true;try{await API.saveAttendance({classId:savedClass,date:savedDate,records});if(version===routeVersion){dirty=false;$('#attendance-status').textContent='ذخیره شد.';toast(APP_CONFIG.demo?'حضور و غیاب در همین مرورگر ذخیره شد.':'حضور و غیاب ذخیره شد.');}}catch(e){toast(e.message);}finally{if(version===routeVersion){btn.disabled=false;$('#attendance-class').disabled=false;$('#attendance-date').disabled=false;$('#all-present').disabled=false;}}};
    await load();
  }
  async function renderDiscipline(version){
    const records=await API.discipline();if(version!==routeVersion)return;
    $('#panel-body').innerHTML=`<div class="surface"><h2>ثبت یادداشت انضباطی</h2><form id="discipline-form"><div class="form-row"><label>کلاس<select id="discipline-class" required>${classOptions()}</select></label><label>دانش‌آموز<select id="discipline-student" required></select></label><label>نوع یادداشت<select name="type"><option value="positive">تشویق</option><option value="attention">نیاز به پیگیری</option></select></label><label>تاریخ<input name="date" type="date" value="${today()}" required></label></div><label class="form-stack">شرح یادداشت<textarea name="note" rows="3" minlength="3" maxlength="1000" placeholder="برای مثال: مشارکت خوب در حل تمرین گروهی" required></textarea></label><div class="save-bar"><button class="btn btn-green" id="discipline-save" type="submit" disabled>ثبت یادداشت</button><span id="discipline-form-status" role="status"></span></div></form></div><div class="surface"><h2>یادداشت‌های ثبت‌شده</h2><div id="discipline-records"></div></div>`;
    let studentLoad=0;
    const loadStudents=async()=>{const op=++studentLoad;$('#discipline-student').disabled=true;$('#discipline-save').disabled=true;try{const data=await API.students($('#discipline-class').value);if(version!==routeVersion||op!==studentLoad)return;$('#discipline-student').innerHTML=data.length?data.map(s=>`<option value="${esc(s.id)}">${esc(s.name)}</option>`).join(''):'<option value="">بدون دانش‌آموز</option>';$('#discipline-student').disabled=false;$('#discipline-save').disabled=!data.length;$('#discipline-form-status').textContent='';}catch(e){if(version===routeVersion)$('#discipline-form-status').textContent=e.message;}};
    $('#discipline-class').onchange=loadStudents;
    const renderRecords=()=>{$('#discipline-records').innerHTML=records.length?records.map(r=>`<article class="record"><div class="record-head"><strong>${esc(r.studentName)}</strong><span class="badge">${r.type==='positive'?'تشویق':'نیاز به پیگیری'}</span></div><small>${esc(classBy(r.classId)?.title||'')} • ${esc(prettyDate(r.date))}</small><p>${esc(r.note)}</p></article>`).join(''):'<p class="help">هنوز یادداشتی ثبت نشده است.</p>';};renderRecords();
    $('#discipline-form').onsubmit=async e=>{e.preventDefault();const f=e.target,d=new FormData(f),note=d.get('note').trim();if(note.length<3){toast('شرح یادداشت باید حداقل ۳ حرف داشته باشد.');return;}const btn=$('#discipline-save');btn.disabled=true;try{const r=await API.addDiscipline({classId:$('#discipline-class').value,studentId:$('#discipline-student').value,studentName:$('#discipline-student').selectedOptions[0].textContent,type:d.get('type'),date:d.get('date'),note});if(version===routeVersion){records.unshift(r);renderRecords();f.elements.note.value='';toast('یادداشت ثبت شد.');}}catch(e){toast(e.message);}finally{if(version===routeVersion)btn.disabled=false;}};
    await loadStudents();
  }
  async function renderSchedule(version){
    schedule=await API.schedule();if(version!==routeVersion)return;
    $('#panel-body').innerHTML=`<div class="surface"><h2>افزودن به برنامه هفتگی</h2><form id="schedule-form"><div class="form-row"><label>کلاس<select name="classId" required>${classOptions()}</select></label><label>روز<select name="day">${dayOptions()}</select></label><label>از ساعت<input name="start" type="time" value="08:00" required></label><label>تا ساعت<input name="end" type="time" value="09:30" required></label><label>موضوع جلسه<input name="topic" maxlength="100" required placeholder="مثلاً حل تمرین تابع"></label><button type="submit" class="btn btn-green">افزودن</button></div></form></div><div id="schedule-table"></div>`;
    const table=()=>{$('#schedule-table').innerHTML=schedule.length?`<div class="table-wrap"><table><thead><tr><th>روز</th><th>کلاس</th><th>ساعت</th><th>موضوع</th><th>عملیات</th></tr></thead><tbody>${[...schedule].sort((a,b)=>a.day-b.day||a.start.localeCompare(b.start)).map(s=>`<tr><td>${days[s.day]}</td><td>${esc(classBy(s.classId)?.title||'')}</td><td>${fa(s.start)} تا ${fa(s.end)}</td><td>${esc(s.topic)}</td><td><button class="danger-link" data-remove="${esc(s.id)}" aria-label="حذف جلسه ${esc(s.topic)}">حذف</button></td></tr>`).join('')}</tbody></table></div>`:'<p class="empty">هنوز برنامه‌ای ثبت نشده است.</p>';};table();
    $('#schedule-form').onsubmit=async e=>{e.preventDefault();const d=new FormData(e.target),body={classId:d.get('classId'),day:Number(d.get('day')),start:d.get('start'),end:d.get('end'),topic:d.get('topic').trim()};if(!body.topic){toast('موضوع جلسه را وارد کنید.');return;}if(body.start>=body.end){toast('ساعت پایان باید بعد از ساعت شروع باشد.');return;}if(schedule.some(s=>s.day===body.day&&s.start<body.end&&body.start<s.end)){toast('این بازه با یکی از کلاس‌های مدرس تداخل دارد.');return;}const btn=$('button[type="submit"]',e.target);btn.disabled=true;try{const item=await API.addSchedule(body);schedule.push(item);if(version===routeVersion){table();toast('جلسه به برنامه اضافه شد.');}}catch(e){toast(e.message);}finally{btn.disabled=false;}};
    $('#schedule-table').onclick=async e=>{const btn=e.target.closest('[data-remove]');if(!btn)return;if(!confirm('این جلسه از برنامه هفتگی حذف شود؟'))return;btn.disabled=true;try{await API.removeSchedule(btn.dataset.remove);schedule=schedule.filter(x=>x.id!==btn.dataset.remove);if(version===routeVersion){table();toast('جلسه حذف شد.');}}catch(e){toast(e.message);btn.disabled=false;}};
  }
  function renderSettings(){$('#panel-body').innerHTML=`<div class="surface"><h2>نام مدرسه</h2><p>این نام در فضای آموزشی نمایش داده می‌شود.</p><form id="school-form" class="form-stack"><label>نام مدرسه<input name="name" value="${esc(school.name)}" minlength="2" maxlength="80" required></label><div><button class="btn btn-green" type="submit">ذخیره تغییرات</button></div></form></div>`;$('#school-form').onsubmit=async e=>{e.preventDefault();const name=new FormData(e.target).get('name').trim();if(name.length<2){toast('نام مدرسه باید حداقل ۲ حرف باشد.');return;}const btn=$('button',e.target);btn.disabled=true;try{school=await API.saveSchool({name});setSchool();toast('نام مدرسه ذخیره شد.');}catch(e){toast(e.message);}finally{btn.disabled=false;}};}
  $('#auth-form').addEventListener('submit',submitAuth);
  $('#demo-enter')?.addEventListener('click',()=>completeLogin({id:'demo',name:'مدرس نمونه',role:'teacher'}));
  $('#logout-button')?.addEventListener('click',logout);
  async function init(){
    if(!APP_CONFIG.demo){$('.subtle-tag').hidden=true;$('.demo-note').hidden=true;$('#mode-label').textContent='متصل به سامانه';}
    try{[classes,schedule,school]=await Promise.all([API.classes(),API.schedule(),API.school()]);setSchool();}catch(e){$('#panel-body').innerHTML=error(e)+'<button class="btn btn-outline" id="reload-data">تلاش دوباره</button>';$('#reload-data').onclick=()=>location.reload();}
    if(!location.hash||location.hash==='#'||location.hash==='#home')location.hash='login';
    window.addEventListener('hashchange',route);
    await route();
  }
  init();
})();
