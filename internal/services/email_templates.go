package services

import (
	"fmt"
	"html"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// EmailTemplate defines the structure for email content
type EmailTemplate struct {
	Subject string
	Body    string
	IsHTML  bool
}

// GetEnrollmentPendingEmailTemplate returns email template for "Em Análise" status
func GetEnrollmentPendingEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, orgaoName string, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Inscrição recebida! - %s", curso.Titulo)
	meusCursosURL := fmt.Sprintf("https://%s/servicos/cursos/meus-cursos", prefrioDomain)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Recebemos seu interesse na atividade <strong>%s</strong>.</p>

    <p>Sua inscrição será avaliada pela equipe responsável do(a) <strong>%s</strong>. Fique atento(a) ao seu e-mail: enviaremos uma atualização assim que o resultado for liberado.</p>

    <p>👉 Se preferir, pode acompanhar também o status clicando em "Meus Cursos" na plataforma <a href="%s" style="color: #0066cc; text-decoration: none;">Oportunidades Cariocas</a>.</p>

    <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(inscricao.Name),
		html.EscapeString(curso.Titulo),
		html.EscapeString(orgaoName),
		html.EscapeString(meusCursosURL),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetEnrollmentApprovedEmailTemplate returns email template for "Inscrito" status
func GetEnrollmentApprovedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, orgaoName string, scheduleInfo *ScheduleInfo, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Parabéns! Sua inscrição foi confirmada - %s", curso.Titulo)

	cursosURL := fmt.Sprintf("https://%s/servicos/cursos", prefrioDomain)

	locationInfoStr := getLocationString(scheduleInfo, curso)
	scheduleInfoStr := getScheduleInfoString(scheduleInfo, curso)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Temos uma ótima notícia! Você está confirmado(a) na atividade <strong>%s</strong>.</p>

    <p><strong>Local e Horário:</strong></p>
    <p>%s<br>
    %s</p>

    <p>A partir de agora, caso haja necessidade de informações específicas sobre a atividade, o contato será realizado diretamente pela equipe responsável do(a) <strong>%s</strong>. Fique atento(a) ao seu e-mail e/ou telefone.</p>

    <p>Em caso de dúvidas, acesse o site e/ou acompanhe as redes sociais do(a) <strong>%s</strong>.</p>

    <p>Aproveite para conferir outros cursos na nossa plataforma <a href="%s" style="color: #0066cc; text-decoration: none;">Oportunidades Cariocas</a>.</p>

    <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(inscricao.Name),
		html.EscapeString(curso.Titulo),
		locationInfoStr,
		scheduleInfoStr,
		html.EscapeString(orgaoName),
		html.EscapeString(orgaoName),
		html.EscapeString(cursosURL),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetEnrollmentRejectedEmailTemplate returns email template for "Recusado" status
func GetEnrollmentRejectedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, orgaoName, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Informações sobre sua inscrição - %s", curso.Titulo)

	cursosURL := fmt.Sprintf("https://%s/servicos/cursos", prefrioDomain)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Agradecemos o seu interesse na atividade <strong>%s</strong>.</p>

    <p>Após análise da instituição responsável, informamos que sua inscrição não foi selecionada desta vez.</p>

    <p>Esclarecemos que a não aprovação pode ocorrer pelo não cumprimento de requisitos específicos ou, ainda, pela limitação do número de vagas, o que impede a seleção de todos os candidatos aptos.</p>

    <p>Se surgir novas vagas, a equipe do(a) <strong>%s</strong> poderá entrar em contato com você. Por isso, sugerimos que fique de olho no seu e-mail e no seu telefone.</p>

    <p>💡 <strong>Mas não desanime!</strong> Convidamos você a conhecer outras oportunidades disponíveis em nossa plataforma. Fique de olho e não perca nenhuma das atividades oferecidas pela Prefeitura do Rio, você não vai querer ficar de fora, né? 😉</p>
    <p>👉 Acesse aqui: <a href="%s" style="color: #0066cc; text-decoration: none;">Oportunidades Cariocas</a></p>

    <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(inscricao.Name),
		html.EscapeString(curso.Titulo),
		html.EscapeString(orgaoName),
		html.EscapeString(cursosURL),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetEnrollmentConcludedEmailTemplate returns email template when a course is concluded
func GetEnrollmentConcludedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Parabéns! 🎉 Certificado disponível - %s", curso.Titulo)
	cursosURL := fmt.Sprintf("https://%s/servicos/cursos", prefrioDomain)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Parabéns! Você concluiu a atividade “<strong>%s</strong>” com sucesso! 🎉</p>

    <p>🎓 Para celebrar sua conquista, seu certificado de conclusão já está disponível. Você pode acessá-lo e baixá-lo diretamente na seção “Certificados” na plataforma do Oportunidades Cariocas.</p>

    <p>👉 Acesse aqui: <a href="%s" style="color: #0066cc; text-decoration: none;">Oportunidades Cariocas</a></p>

    <p>Estamos aqui na torcida pelo seu sucesso e esperamos você nas próximas atividades! 👋</p>

    <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(inscricao.Name),
		html.EscapeString(curso.Titulo),
		html.EscapeString(cursosURL),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetEnrollmentClassReminderEmailTemplate returns email template for D-1 class reminder
func GetEnrollmentClassReminderEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, orgaoName string, scheduleInfo *ScheduleInfo, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Atenção, %s! Falta pouco para a atividade %s", inscricao.Name, curso.Titulo)

	locationInfoStr := getLocationString(scheduleInfo, curso)
	scheduleInfoStr := getScheduleInfoString(scheduleInfo, curso)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Temos um lembrete importante para você! A atividade de <strong>%s</strong> começa amanhã! 🎉</p>

    <p>Confira abaixo as informações da atividade para você não perder nenhum detalhe:</p>
    <p>%s<br>
    %s</p>

    <p>Caso tenha alguma informação extra sobre a atividade (como material, link de acesso ou instruções), a equipe responsável do(a) <strong>%s</strong> poderá entrar em contato direto com você por e-mail ou telefone. Fique de olho!</p>

    <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(inscricao.Name),
		html.EscapeString(curso.Titulo),
		locationInfoStr,
		scheduleInfoStr,
		html.EscapeString(orgaoName),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetScheduleChangedEmailTemplate returns email template for schedule changes
func GetScheduleChangedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, scheduleInfo *ScheduleInfo, orgaoName, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Troca de turma realizada com sucesso - %s", curso.Titulo)

	cursosURL := fmt.Sprintf("https://%s/servicos/cursos", prefrioDomain)

	locationInfoStr := getLocationString(scheduleInfo, curso)
	scheduleInfoStr := getScheduleInfoString(scheduleInfo, curso)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Sua troca de turma na atividade <strong>%s</strong> foi realizada com sucesso. ✅</p>

		<p><strong>Fique atento(a) ao seu novo horário:</strong></p>
    <p>%s<br>
    %s</p>

    <p>A partir de agora, caso haja necessidade de informações específicas sobre a atividade, o contato será realizado diretamente pela equipe responsável do(a) <strong>%s</strong>. Fique atento(a) ao seu e-mail e/ou telefone.</p>

    <p>Em caso de dúvidas, acesse o site e/ou acompanhe as redes sociais do(a) <strong>%s</strong>.</p>

    <p>Aproveite para conferir outros cursos na nossa plataforma <a href="%s" style="color: #0066cc; text-decoration: none;">Oportunidades Cariocas</a>.</p>

    <p>Até a próxima 👋</p>

    <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(inscricao.Name),
		html.EscapeString(curso.Titulo),
		locationInfoStr,
		scheduleInfoStr,
		html.EscapeString(orgaoName),
		html.EscapeString(orgaoName),
		html.EscapeString(cursosURL),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetCandidaturaEnviadaEmailTemplate returns email template for received applications
func GetCandidaturaEnviadaEmailTemplate(candidatura *empregabilidade.Candidatura, vaga *empregabilidade.Vaga, empresa *empregabilidade.Empresa, orgaoName, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Candidatura recebida! - %s", vaga.Titulo)
	candidaturasURL := fmt.Sprintf("https://%s/servicos/trabalho/minhas-candidaturas", prefrioDomain)

	nome := ""
	if candidatura.Nome != nil {
		nome = *candidatura.Nome
	}

	var companyName string
	if empresa != nil {
		if empresa.NomeFantasia != "" {
			companyName = empresa.NomeFantasia
		} else {
			companyName = empresa.RazaoSocial
		}
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Recebemos sua inscrição para a vaga de %s na empresa %s, em parceria com %s.</p>

		<p>Fique atento(a) ao seu e-mail, pois enviaremos atualizações sobre o andamento do processo seletivo em breve.</p>

    <p>👉 Você pode acompanhar o status da sua inscrição a qualquer momento na seção: <a href="%s" style="color: #0066cc; text-decoration: none;">Minhas candidaturas</a>.</p>

    <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(nome),
		html.EscapeString(vaga.Titulo),
		html.EscapeString(companyName),
		html.EscapeString(orgaoName),
		html.EscapeString(candidaturasURL),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetCandidaturaAprovadaEmailTemplate returns email template for approved applications
func GetCandidaturaAprovadaEmailTemplate(candidatura *empregabilidade.Candidatura, vaga *empregabilidade.Vaga, empresa *empregabilidade.Empresa) EmailTemplate {
	subject := fmt.Sprintf(`Parabéns! 🎉Candidatura aprovada - "%s"`, vaga.Titulo)

	var companyName string

	if empresa.NomeFantasia != "" {
		companyName = empresa.NomeFantasia
	} else {
		companyName = empresa.RazaoSocial
	}

	nome := ""
	if candidatura.Nome != nil {
		nome = *candidatura.Nome
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Estamos muito felizes em informar que você foi aprovado (a) para a vaga “%s”. Parabéns por essa conquista! 🥳👏</p>

		<p>💡 Fique atento(a) ao seu e-mail/ telefone, pois a empresa “%s” entrará em contato em breve para alinhar os detalhes da contratação.</p>

    <p>Desejamos muito sucesso em seu novo emprego! 🎉🎉</p>

     <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(nome),
		html.EscapeString(vaga.Titulo),
		html.EscapeString(companyName),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetCandidaturaReprovadaEmailTemplate returns email template for failed applications
func GetCandidaturaReprovadaEmailTemplate(candidatura *empregabilidade.Candidatura, vaga *empregabilidade.Vaga, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf(`Informações sobre sua candidatura - %s`, vaga.Titulo)
	vagasURL := fmt.Sprintf("https://%s/servicos/trabalho", prefrioDomain)

	nome := ""
	if candidatura.Nome != nil {
		nome = *candidatura.Nome
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

		<p>Agradecemos o seu interesse na vaga de “%s”.</p>

    <p>Após a análise do seu perfil em relação aos requisitos da vaga, informamos que sua candidatura não seguirá para as próximas etapas desta vez.</p>

		<p>💡<strong>Não desanime!</strong> Fique de olho e se inscreva nas próximas oportunidades que combinem com o seu perfil. A vaga certa pode estar à sua espera! 😉</p>

    <p>👉 Acesse aqui as vagas disponíveis: <a href="%s" style="color: #0066cc; text-decoration: none;">%s</a></p>

     <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(nome),
		html.EscapeString(vaga.Titulo),
		html.EscapeString(vagasURL),
		html.EscapeString(vagasURL),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetCandidaturaProximaEtapaEmailTemplate returns email when a candidate advances to the next selective stage
func GetCandidaturaProximaEtapaEmailTemplate(candidatura *empregabilidade.Candidatura, vaga *empregabilidade.Vaga, etapaNome, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Boas notícias! Você avançou no processo para vaga de %s 🚀", vaga.Titulo)
	candidaturasURL := fmt.Sprintf("https://%s/servicos/trabalho/minhas-candidaturas", prefrioDomain)

	nome := ""
	if candidatura.Nome != nil {
		nome = *candidatura.Nome
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Temos uma ótima notícia: sua inscrição para a vaga de “<strong>%s</strong>” foi selecionada para a próxima etapa “<strong>%s</strong>”! 🎉</p>

    <p>💡 <strong>O que acontece agora?</strong><br>
    Fique de olho no seu e-mail e no seu celular. A equipe responsável pode entrar em contato com você em breve para o andamento dos próximos passos.</p>

    <p>Você também pode acompanhar tudo pela nossa plataforma.</p>

    <p>👉 Acesse aqui: <a href="%s" style="color: #0066cc; text-decoration: none;">Minhas candidaturas</a></p>

    <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		html.EscapeString(nome),
		html.EscapeString(vaga.Titulo),
		html.EscapeString(etapaNome),
		html.EscapeString(candidaturasURL),
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// Build location info from schedule or course (HTML-safe for email bodies).
func getLocationString(scheduleInfo *ScheduleInfo, curso *models.Curso) string {
	var locationInfoStr string
	if scheduleInfo != nil && scheduleInfo.Address != "" {
		if scheduleInfo.Address == "online" {
			locationInfoStr = "📍 Endereço: online"
		} else {
			locationInfoStr = fmt.Sprintf("📍 Endereço: %s", html.EscapeString(scheduleInfo.Address))
		}
	} else if curso.Modalidade == models.ModalidadePresencial || curso.Modalidade == models.ModalidadePresencialLegacy {
		if curso.LocalRealizacao != "" {
			locationInfoStr = fmt.Sprintf("📍 Endereço: %s", html.EscapeString(curso.LocalRealizacao))
		} else {
			locationInfoStr = "📍 Endereço: a confirmar"
		}
	} else {
		locationInfoStr = "📍 Endereço: online"
	}
	return locationInfoStr
}

// Build schedule info from enrollment's schedule or fallback to course data (HTML-safe).
func getScheduleInfoString(scheduleInfo *ScheduleInfo, curso *models.Curso) string {
	var scheduleInfoStr string
	if scheduleInfo != nil && scheduleInfo.ClassStartDate != "" {
		if scheduleInfo.ClassTime != "" {
			scheduleInfoStr = fmt.Sprintf("⏰ Horário de início: %s às %s",
				html.EscapeString(scheduleInfo.ClassStartDate),
				html.EscapeString(scheduleInfo.ClassTime))
		} else {
			scheduleInfoStr = fmt.Sprintf("⏰ Data de início: %s", html.EscapeString(scheduleInfo.ClassStartDate))
		}
		if scheduleInfo.ClassDays != "" {
			scheduleInfoStr += fmt.Sprintf(" (%s)", html.EscapeString(scheduleInfo.ClassDays))
		}
	} else if curso.DataInicio != nil {
		scheduleInfoStr = fmt.Sprintf("⏰ Horário de início: %s", html.EscapeString(curso.DataInicio.Format("02/01/2006 15:04")))
	} else {
		scheduleInfoStr = "⏰ Horário de início: a confirmar"
	}
	return scheduleInfoStr
}
