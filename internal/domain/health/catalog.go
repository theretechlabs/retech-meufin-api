package health

import "github.com/retechfin/retechfin-api/internal/domain/catalog"

// Catálogos do módulo de saúde publicados em GET /api/v1/catalogs
// (docs/catalogs.md). Novo valor num enum = nova constante + rótulo aqui.

var relationshipOrder = []Relationship{
	RelationshipSelf, RelationshipSpouse, RelationshipChild, RelationshipParent, RelationshipOther,
}

var relationshipLabels = map[Relationship]string{
	RelationshipSelf:   "Eu",
	RelationshipSpouse: "Cônjuge",
	RelationshipChild:  "Filho(a)",
	RelationshipParent: "Pai/Mãe",
	RelationshipOther:  "Outro",
}

// genderItems: `Gender` do membro é texto livre no banco; o catálogo é a lista
// curada que o admin oferece.
var genderItems = []catalog.Item{
	{Slug: "male", Name: "Masculino"},
	{Slug: "female", Name: "Feminino"},
	{Slug: "other", Name: "Outro"},
}

var labKindOrder = []LabKind{
	LabKindLaboratorio, LabKindClinica, LabKindHospital, LabKindConsultorio, LabKindOtica, LabKindOutros,
}

var labKindLabels = map[LabKind]string{
	LabKindLaboratorio: "Laboratório",
	LabKindClinica:     "Clínica",
	LabKindHospital:    "Hospital",
	LabKindConsultorio: "Consultório",
	LabKindOtica:       "Ótica",
	LabKindOutros:      "Outros",
}

var planTypeOrder = []PlanType{PlanTypeIndividual, PlanTypeFamiliar, PlanTypeEmpresarial, PlanTypeOdontologico}

var planTypeLabels = map[PlanType]string{
	PlanTypeIndividual:   "Individual",
	PlanTypeFamiliar:     "Familiar",
	PlanTypeEmpresarial:  "Empresarial",
	PlanTypeOdontologico: "Odontológico",
}

var appointmentKindOrder = []AppointmentKind{
	AppointmentKindConsulta, AppointmentKindExame, AppointmentKindRetorno, AppointmentKindVacina,
	AppointmentKindProcedimento, AppointmentKindTerapia, AppointmentKindOdontologia,
}

var appointmentKindLabels = map[AppointmentKind]string{
	AppointmentKindConsulta:     "Consulta",
	AppointmentKindExame:        "Exame",
	AppointmentKindRetorno:      "Retorno",
	AppointmentKindVacina:       "Vacina",
	AppointmentKindProcedimento: "Procedimento",
	AppointmentKindTerapia:      "Terapia",
	AppointmentKindOdontologia:  "Odontologia",
}

var specialtyOrder = []Specialty{
	SpecialtyClinicaGeral, SpecialtyCardiologia, SpecialtyPediatria, SpecialtyGinecologia,
	SpecialtyDermatologia, SpecialtyOrtopedia, SpecialtyOftalmologia, SpecialtyOtorrino,
	SpecialtyPsicologia, SpecialtyPsiquiatria, SpecialtyNutricao, SpecialtyEndocrinologia,
	SpecialtyUrologia, SpecialtyGeriatria, SpecialtyOdontologia, SpecialtyFisioterapia, SpecialtyOutros,
}

var specialtyLabels = map[Specialty]string{
	SpecialtyClinicaGeral:   "Clínica geral",
	SpecialtyCardiologia:    "Cardiologia",
	SpecialtyPediatria:      "Pediatria",
	SpecialtyGinecologia:    "Ginecologia",
	SpecialtyDermatologia:   "Dermatologia",
	SpecialtyOrtopedia:      "Ortopedia",
	SpecialtyOftalmologia:   "Oftalmologia",
	SpecialtyOtorrino:       "Otorrinolaringologia",
	SpecialtyPsicologia:     "Psicologia",
	SpecialtyPsiquiatria:    "Psiquiatria",
	SpecialtyNutricao:       "Nutrição",
	SpecialtyEndocrinologia: "Endocrinologia",
	SpecialtyUrologia:       "Urologia",
	SpecialtyGeriatria:      "Geriatria",
	SpecialtyOdontologia:    "Odontologia",
	SpecialtyFisioterapia:   "Fisioterapia",
	SpecialtyOutros:         "Outros",
}

var appointmentStatusOrder = []AppointmentStatus{
	AppointmentStatusAgendada, AppointmentStatusConfirmada, AppointmentStatusRealizada,
	AppointmentStatusCancelada, AppointmentStatusFaltou,
}

var appointmentStatusLabels = map[AppointmentStatus]string{
	AppointmentStatusAgendada:   "Agendada",
	AppointmentStatusConfirmada: "Confirmada",
	AppointmentStatusRealizada:  "Realizada",
	AppointmentStatusCancelada:  "Cancelada",
	AppointmentStatusFaltou:     "Faltou",
}

var memberDocTypeOrder = []MemberDocType{
	MemberDocCPF, MemberDocRG, MemberDocCNH, MemberDocPassaporte, MemberDocCarteiraTrabalho,
	MemberDocCertidaoNasc, MemberDocTituloEleitor, MemberDocCartaoSUS, MemberDocPlanoSaude, MemberDocOutro,
}

var memberDocTypeLabels = map[MemberDocType]string{
	MemberDocCPF:              "CPF",
	MemberDocRG:               "RG",
	MemberDocCNH:              "CNH",
	MemberDocPassaporte:       "Passaporte",
	MemberDocCarteiraTrabalho: "Carteira de Trabalho",
	MemberDocCertidaoNasc:     "Certidão de Nascimento",
	MemberDocTituloEleitor:    "Título de Eleitor",
	MemberDocCartaoSUS:        "Cartão SUS",
	MemberDocPlanoSaude:       "Plano de Saúde",
	MemberDocOutro:            "Outro (informe o rótulo)",
}

var planDocTypeOrder = []PlanDocType{
	PlanDocContrato, PlanDocCarteirinha, PlanDocManual, PlanDocTabelaCoparticipacao, PlanDocAditivoReajuste,
	PlanDocBoleto, PlanDocTermoAdesao, PlanDocComprovanteCarencia, PlanDocFormularioReembolso,
	PlanDocDeclaracaoIR, PlanDocRedeCredenciada, PlanDocLaudo, PlanDocOutro,
}

var planDocTypeLabels = map[PlanDocType]string{
	PlanDocContrato:             "Contrato / Apólice",
	PlanDocCarteirinha:          "Carteirinha",
	PlanDocManual:               "Manual do Beneficiário",
	PlanDocTabelaCoparticipacao: "Tabela de Coparticipação",
	PlanDocAditivoReajuste:      "Aditivo / Reajuste",
	PlanDocBoleto:               "Boleto / Comprovante de Pagamento",
	PlanDocTermoAdesao:          "Termo de Adesão",
	PlanDocComprovanteCarencia:  "Comprovante de Carência",
	PlanDocFormularioReembolso:  "Formulário de Reembolso",
	PlanDocDeclaracaoIR:         "Declaração para IR",
	PlanDocRedeCredenciada:      "Rede Credenciada",
	PlanDocLaudo:                "Laudo / Relatório Médico",
	PlanDocOutro:                "Outro (informe o rótulo)",
}

var documentTypeOrder = []DocumentType{
	DocTypeExamResult, DocTypeExamRequest, DocTypeImageReport, DocTypeMedicalReport, DocTypePrescription, DocTypeOther,
}

var documentTypeLabels = map[DocumentType]string{
	DocTypeExamResult:    "Resultado de exame",
	DocTypeExamRequest:   "Solicitação",
	DocTypeImageReport:   "Exame de imagem",
	DocTypeMedicalReport: "Laudo médico",
	DocTypePrescription:  "Receita",
	DocTypeOther:         "Outro",
}

// Origem do resultado de exame: `manual` é o único valor que o usuário escolhe;
// os demais são gravados pela importação de documentos.
var sourceTypeOrder = []SourceType{SourceManual, SourcePDF, SourceImage, SourceOCR, SourceLLM}

var sourceTypeLabels = map[SourceType]string{
	SourceManual: "Manual",
	SourcePDF:    "PDF",
	SourceImage:  "Imagem",
	SourceOCR:    "OCR",
	SourceLLM:    "Extração por IA",
}

var examResultStatusOrder = []ExamResultStatus{
	ResultStatusDraft, ResultStatusProcessing, ResultStatusExtracted, ResultStatusReviewed, ResultStatusFailed,
}

var examResultStatusLabels = map[ExamResultStatus]string{
	ResultStatusDraft:      "Rascunho",
	ResultStatusProcessing: "Processando",
	ResultStatusExtracted:  "Extraído",
	ResultStatusReviewed:   "Revisado",
	ResultStatusFailed:     "Falhou",
}

// markerCategoryItems: `Category` do marcador é texto livre no banco (seed em
// PT); o catálogo é a lista curada que o admin oferece.
var markerCategoryItems = []catalog.Item{
	{Slug: "hematologia", Name: "Hematologia"},
	{Slug: "bioquimica", Name: "Bioquímica"},
	{Slug: "lipidico", Name: "Perfil lipídico"},
	{Slug: "hepatico", Name: "Hepático"},
	{Slug: "renal", Name: "Renal"},
	{Slug: "eletrolitos", Name: "Eletrólitos"},
	{Slug: "hormonios", Name: "Hormônios"},
	{Slug: "vitaminas", Name: "Vitaminas"},
	{Slug: "inflamacao", Name: "Inflamação"},
	{Slug: "outros", Name: "Outros"},
}

// Catalogs devolve os catálogos do módulo de saúde, na ordem de exibição.
func Catalogs() []catalog.Catalog {
	return []catalog.Catalog{
		{Key: "health.relationship", Items: catalog.FromLabels(relationshipOrder, relationshipLabels)},
		{Key: "health.gender", Items: genderItems},
		{Key: "health.lab_kind", Items: catalog.FromLabels(labKindOrder, labKindLabels)},
		{Key: "health.plan_type", Items: catalog.FromLabels(planTypeOrder, planTypeLabels)},
		{Key: "health.appointment_kind", Items: catalog.FromLabels(appointmentKindOrder, appointmentKindLabels)},
		{Key: "health.specialty", Items: catalog.FromLabels(specialtyOrder, specialtyLabels)},
		{Key: "health.appointment_status", Items: catalog.FromLabels(appointmentStatusOrder, appointmentStatusLabels)},
		{Key: "health.member_doc_type", Items: catalog.FromLabels(memberDocTypeOrder, memberDocTypeLabels)},
		{Key: "health.plan_doc_type", Items: catalog.FromLabels(planDocTypeOrder, planDocTypeLabels)},
		{Key: "health.document_type", Items: catalog.FromLabels(documentTypeOrder, documentTypeLabels)},
		{Key: "health.exam_source_type", Items: catalog.FromLabels(sourceTypeOrder, sourceTypeLabels)},
		{Key: "health.exam_result_status", Items: catalog.FromLabels(examResultStatusOrder, examResultStatusLabels)},
		{Key: "health.marker_category", Items: markerCategoryItems},
	}
}
