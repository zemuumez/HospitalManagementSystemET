"""Reproducible, read-only inventory of Laravel presentation fields for the preview."""
import json, re, shutil
from pathlib import Path
root = Path(__file__).resolve().parents[1]
views = root / 'review/legacy/resources/views'
groups = {
 'Patient ID Card': 'patient_id_card_template generate_patient_id_card',
 'Users': 'users admins accountants nurses lab_technicians receptionists pharmacists',
 'Appointments': 'appointments appointment_calendars appointment_transaction patient_queues',
 'IPD / OPD': 'ipd_patient_departments opd_patient_departments',
 'Billing': 'accounts employee_payrolls invoices payments payment_reports advanced_payments bills manual_bill_payments',
 'Bed Management': 'bed_status bed_assigns beds bed_types',
 'Blood Bank': 'blood_banks blood_donors blood_donations blood_issues',
 'Documents': 'documents document_types',
 'Doctors': 'doctors doctor_departments schedules doctor_holiday lunch_breaks',
 'Prescriptions': 'prescriptions',
 'Diagnosis': 'patient_diagnosis_test diagnosis_categories',
 'Enquiries': 'enquiries',
 'Finance': 'incomes expenses',
 'Front Office': 'call_logs visitors postals complaints',
 'Front CMS': 'front_settings services notice_boards testimonials',
 'Hospital Charges': 'charge_categories charges doctor_opd_charges',
 'Inventory': 'item_categories items item_stocks issued_items',
 'Live Consultations': 'live_consultations live_consultations/live_meetings goole_meet_consultation',
 'Medicine': 'categories brands medicines purchase-medicines used-medicine medicine-bills',
 'Patients': 'patients patient_cases case_handlers patient_admissions',
 'Pathology': 'pathology_categories pathology_units pathology_parameter pathology_tests',
 'Reports': 'birth_reports death_reports investigation_reports operation_reports',
 'Radiology': 'radiology_categories radiology_tests',
 'Services': 'insurances packages ambulances ambulance_calls',
 'SMS / Mail': 'sms emails email-template',
 'Settings': 'settings hospital_schedule currency_settings operation_categories operations payment_gateway add_custom_fields',
 'Vaccinations': 'vaccinations vaccinated_patients',
 'Odontogram': 'odontogram',
 'Add-ons': 'add-on',
 'Clinical Records': 'ipd_diagnoses ipd_consultant_registers ipd_operation ipd_charges ipd_prescriptions ipd_timelines ipd_payments ipd_bills opd_diagnoses opd_prescriptions opd_timelines',
}
names = {'patient_id_card_template':'Patient ID Card Templates','generate_patient_id_card':'Generate Patient ID Card','ipd_patient_departments':'IPD Patients','opd_patient_departments':'OPD Patients','categories':'Medicine Categories','brands':'Medicine Brands','goole_meet_consultation':'Google Meet Consultations','postals':'Postal Receive / Dispatch','patient_diagnosis_test':'Diagnosis Tests','front_settings':'Front Settings','settings':'General Settings','currency_settings':'Currencies','add_custom_fields':'Custom Fields','manual_bill_payments':'Manual Payment Approvals','doctor_holiday':'Doctor Holidays','pathology_parameter':'Pathology Parameters','used-medicine':'Used Medicines','medicine-bills':'Medicine Bills','purchase-medicines':'Medicine Purchases'}
def human(s):
    return re.sub(r'(?<=[a-z])(?=[A-Z])',' ',s.replace('_',' ').replace('-',' ')).title().replace('Ipd','IPD').replace('Opd','OPD').replace('Sms','SMS').replace('Dob','Date of Birth').replace(' Id',' ID').replace('Url','URL')
def clean(s): return re.sub(r'\{\{--.*?--\}\}|<!--.*?-->', '', s, flags=re.S)
def fields_for(folder):
    files = [p for p in folder.rglob('*.blade.php') if p.name in ['fields.blade.php','create_modal.blade.php','modal.blade.php','create.blade.php','general.blade.php','edit.blade.php','add_modal.blade.php','medicine-table.blade.php'] or 'create' in p.name or p.name.startswith('add.')]
    found=[]
    for file in files:
        body=clean(file.read_text(encoding='utf-8'))
        for m in re.finditer(r"Form::label\(\s*'([^']+)'\s*,\s*(?:__\(\s*'([^']+)'\s*\)|'([^']+)')",body):
            key=m[1]
            if any(f['key']==key for f in found): continue
            label=human((m[2] or m[3]).split('.')[-1]).rstrip(':')
            segment=body[m.end():body.find('Form::label',m.end()) if body.find('Form::label',m.end())!=-1 else len(body)]
            inp=re.search(r"Form::(text|textarea|select|number|date|password|email|tel|radio|checkbox|file|color)\(\s*'([^']+)'",segment)
            kind=inp[1] if inp else 'text'
            if inp: key=inp[2]
            if any(f['key']==key for f in found): continue
            if kind=='text' and re.search('date|dob',key): kind='date'
            if key=='email': kind='email'
            if kind in ['checkbox','radio']: kind='select'
            if key in ['status','is_active']: kind='select'
            if key in ['image','file','document','attachment']: kind='file'
            label=re.sub(r'(Address)([12])',r'\1 \2',label)
            found.append({'key':key,'label':label,'type':kind,'required':bool(re.search(r"'required'|class=\"required\"",segment)), 'source':str(file.relative_to(root)).replace('\\','/')})
        for m in re.finditer(r"Form::(text|textarea|select|number|date|password|email|tel|file)\(\s*'([^']+)'",body):
            key=m[2]
            if any(f['key']==key for f in found) or '[]' in key: continue
            segment=body[m.end():body.find('}}',m.end())]
            kind=m[1]
            if re.search('date|dob',key): kind='date'
            found.append({'key':key,'label':human(key),'type':kind,'required':"'required'" in segment,'source':str(file.relative_to(root)).replace('\\','/')})
    return found
tables=list((root/'review/legacy/app/Livewire').glob('*Table.php'))
catalog=[]
for group, folders in groups.items():
    for folder in folders.split():
        path=views/folder
        fields=fields_for(path)
        target=re.sub('[^a-z]','',folder.lower()).rstrip('s')
        table=next((p for p in tables if re.sub('[^a-z]','',p.stem.lower().replace('table','')).rstrip('s')==target),None)
        columns=[]
        if table:
            columns=list(dict.fromkeys(human(x.split('.')[-1]) for x in re.findall(r"Column::make\(__\('([^']+)'\)",clean(table.read_text(encoding='utf-8')))))
            columns=[x for x in columns if x not in ['Action','Created At']]
        if not columns: columns=[f['label'] for f in fields if f['type'] not in ['password','textarea','file']][:5]
        if not fields: fields=[{'key':'name','label':'Name','type':'text','required':True},{'key':'description','label':'Description','type':'textarea','required':False}]
        catalog.append({'id':folder.replace('_','-').replace('/','-'),'title':names.get(folder,human(folder.split('/')[-1])),'group':group,'fields':fields,'columns':columns or ['Name','Description'],'source':str(path.relative_to(root)).replace('\\','/'),'tableSource':str(table.relative_to(root)).replace('\\','/') if table else None,'extracted':bool(fields[0].get('source'))})
out=root/'apps/web/src/lib/legacy-catalog.json'
out.write_text(json.dumps(catalog,indent=2)+'\n',encoding='utf-8')
dest=root/'apps/web/public/legacy';dest.mkdir(parents=True,exist_ok=True)
for name in ['Poppins-Regular.ttf','Poppins-Medium.ttf','Poppins-Bold.ttf']:
    shutil.copyfile(root/'review/legacy/resources/assets/theme/fonts'/name,dest/name)
shutil.copyfile(root/'review/legacy/resources/assets/theme/images/logo.png',dest/'logo.png')
print(f'{len(catalog)} screens; {sum(len(x["fields"]) for x in catalog)} fields; {sum(x["extracted"] for x in catalog)} source-backed forms')
report=['# Frontend source coverage','', 'This inventory records extracted presentation fields, not completed behavioral parity. Generic fallback forms and special workflows still need individual verification.','', '| Screen | Form fields | Table source |', '| --- | --- | --- |']
report += [f'| {x["group"]} / {x["title"]} | {len(x["fields"])} ({"extracted" if x["extracted"] else "fallback"}) | {x["tableSource"] or "needs verification"} |' for x in catalog]
(root/'docs/frontend-coverage.md').write_text('\n'.join(report)+'\n',encoding='utf-8')
