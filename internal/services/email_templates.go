package services

import (
	"fmt"

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

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Tudo bem? Recebemos seu interesse na atividade <strong>%s</strong>.</p>

    <p>Neste momento, sua inscrição está sendo avaliada pela equipe responsável do(a) <strong>%s</strong>. Fique atento ao seu e-mail: enviaremos uma atualização assim que o resultado da análise for liberado.</p>

    <p>👉 Se preferir, acompanhe o status clicando em "Meus Cursos" na plataforma Oportunidades Cariocas.</p>

    <p>Boa sorte!</p>

    <p>Equipe do Oportunidades Cariocas</p>

    <p><em>Observação: Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		inscricao.Name,
		curso.Titulo,
		orgaoName,
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

    <p>Temos uma ótima notícia! 🎉 Você está confirmado(a) na atividade <strong>%s</strong>.</p>

    <p><strong>Local e Horário:</strong></p>
    <p>%s<br>
    %s</p>

    <p>A partir de agora, para informações específicas sobre a atividade, o contato será realizado diretamente pela equipe responsável do(a) <strong>%s</strong>. Fique atento(a) ao seu e-mail e/ou telefone.</p>

    <p>Em caso de dúvidas e sugestões, acesse o site e/ou acompanhe as redes sociais do(a) <strong>%s</strong>.</p>

    <p>Aproveite para conferir outros cursos na nossa plataforma <a href="%s" style="color: #0066cc; text-decoration: none;">Oportunidades Cariocas</a>.</p>

    <p>Até a próxima 👋<br>
    Equipe do Oportunidades Cariocas</p>

    <p><em>Observação: Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		inscricao.Name,
		curso.Titulo,
		locationInfoStr,
		scheduleInfoStr,
		orgaoName,
		orgaoName,
		cursosURL,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetEnrollmentRejectedEmailTemplate returns email template for "Recusado" status
func GetEnrollmentRejectedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, prefrioDomain string) EmailTemplate {
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

    <p>Após a análise realizada pela instituição responsável, infelizmente, sua inscrição não foi aprovada desta vez.</p>

    <p>A não aprovação ocorreu por não ter sido cumprido(s) algum(ns) dos requisitos ou critérios definidos para a atividade.</p>

    <p>💡 Mas não desanime! Convidamos você a conhecer outras oportunidades disponíveis em nossa plataforma. Fique de olho e não perca nenhuma das atividades oferecidas pela Prefeitura do Rio, você não vai querer ficar de fora, né? 😉</p>

    <p>👉 Link de acesso: <a href="%s" style="color: #0066cc; text-decoration: none;">Oportunidades Cariocas</a></p>

    <p>Estamos aqui na torcida pelo seu sucesso! Te esperamos por lá! 👋</p>

    <p>Até a próxima 👋<br>
    Equipe do Oportunidades Cariocas</p>

    <p><em>Observação: Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		inscricao.Name,
		curso.Titulo,
		cursosURL,
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

    <p><em>Observação: Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		inscricao.Name,
		curso.Titulo,
		locationInfoStr,
		scheduleInfoStr,
		orgaoName,
		orgaoName,
		cursosURL,
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
	vagasURL := fmt.Sprintf("https://%s/servicos/trabalho", prefrioDomain)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Recebemos sua inscrição para a vaga de %s na empresa %s, em parceria com %s.</p>

		<p>Fique atento(a) ao seu e-mail, pois enviaremos atualizações sobre o andamento do processo seletivo em breve.</p>

    <p>👉 Você pode acompanhar o status da sua inscrição a qualquer momento na seção <a href="%s" style="color: #0066cc; text-decoration: none;">%s</a></p>

    <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		*candidatura.Nome,
		vaga.Titulo,
		empresa.NomeFantasia,
		orgaoName,
		vagasURL,
		vagasURL,
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
		*candidatura.Nome,
		vaga.Titulo,
		empresa.NomeFantasia,
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

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

		>p>Agradecemos o seu interesse na vaga de “%s”.</p>

    <p>Após a análise do seu perfil em relação aos requisitos da vaga, informamos que sua candidatura não seguirá para as próximas etapas desta vez.</p>

		<p>💡<strong>Não desanime!</strong> Fique de olho e se inscreva nas próximas oportunidades que combinem com o seu perfil. A vaga certa pode estar à sua espera! 😉</p>

    <p>👉 Acesse aqui as vagas disponíveis: <a href="%s" style="color: #0066cc; text-decoration: none;">%s</a></p>

     <p><em><strong>Observação:</strong> Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		*candidatura.Nome,
		vaga.Titulo,
		vagasURL,
		vagasURL,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// Build location info from schedule or course
func getLocationString(scheduleInfo *ScheduleInfo, curso *models.Curso) string {
	var locationInfoStr string
	if scheduleInfo != nil && scheduleInfo.Address != "" {
		locationInfoStr = fmt.Sprintf("📍 Endereço: %s", scheduleInfo.Address)
	} else if curso.Modalidade == models.ModalidadePresencial || curso.Modalidade == models.ModalidadePresencialLegacy {
		if curso.LocalRealizacao != "" {
			locationInfoStr = fmt.Sprintf("📍 Endereço: %s", curso.LocalRealizacao)
		} else {
			locationInfoStr = "📍 Endereço: a confirmar"
		}
	} else {
		locationInfoStr = "📍 Endereço: online"
	}
	return locationInfoStr
}

// Build schedule info from enrollment's schedule or fallback to course data
func getScheduleInfoString(scheduleInfo *ScheduleInfo, curso *models.Curso) string {
	var scheduleInfoStr string
	if scheduleInfo != nil && scheduleInfo.ClassStartDate != "" {
		if scheduleInfo.ClassTime != "" {
			scheduleInfoStr = fmt.Sprintf("🕒 Horário de início: %s às %s", scheduleInfo.ClassStartDate, scheduleInfo.ClassTime)
		} else {
			scheduleInfoStr = fmt.Sprintf("🕒 Data de início: %s", scheduleInfo.ClassStartDate)
		}
		if scheduleInfo.ClassDays != "" {
			scheduleInfoStr += fmt.Sprintf(" (%s)", scheduleInfo.ClassDays)
		}
	} else if curso.DataInicio != nil {
		scheduleInfoStr = fmt.Sprintf("🕒 Horário de início: %s", curso.DataInicio.Format("02/01/2006 15:04"))
	} else {
		scheduleInfoStr = "🕒 Horário de início: a confirmar"
	}
	return scheduleInfoStr
}
