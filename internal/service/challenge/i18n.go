package challenge

// I18n provides localized strings for challenge UI.
type I18n struct {
	translations map[string]map[string]string
}

func NewI18n() *I18n {
	i := &I18n{translations: make(map[string]map[string]string)}
	i.loadDefaults()
	return i
}

func (i *I18n) Get(lang, key string) string {
	if t, ok := i.translations[lang]; ok {
		if v, ok := t[key]; ok {
			return v
		}
	}
	// Fallback to English
	if t, ok := i.translations["en"]; ok {
		if v, ok := t[key]; ok {
			return v
		}
	}
	return key
}

func (i *I18n) GetAll(lang string) map[string]string {
	result := make(map[string]string)
	// Start with English defaults
	if en, ok := i.translations["en"]; ok {
		for k, v := range en {
			result[k] = v
		}
	}
	// Override with requested language
	if t, ok := i.translations[lang]; ok {
		for k, v := range t {
			result[k] = v
		}
	}
	return result
}

func (i *I18n) SupportedLanguages() []string {
	langs := make([]string, 0, len(i.translations))
	for k := range i.translations {
		langs = append(langs, k)
	}
	return langs
}

func (i *I18n) loadDefaults() {
	i.translations["en"] = map[string]string{
		"slider.title":       "Drag the slider to verify",
		"slider.success":     "Verification successful",
		"slider.fail":        "Please try again",
		"click.title":        "Click the items in order",
		"click.success":      "Verification successful",
		"click.fail":         "Incorrect order, please try again",
		"puzzle.title":       "Drag the piece to complete the puzzle",
		"loading":            "Loading...",
		"error":              "An error occurred. Please try again.",
		"protected_by":       "Protected by EngageLab CAPTCHA",
		"accessibility.alt":  "Security verification challenge",
		"retry":              "Retry",
		"close":              "Close",
	}

	i.translations["zh"] = map[string]string{
		"slider.title":       "拖动滑块完成验证",
		"slider.success":     "验证成功",
		"slider.fail":        "请重试",
		"click.title":        "请按顺序点击目标",
		"click.success":      "验证成功",
		"click.fail":         "顺序错误，请重试",
		"puzzle.title":       "拖动拼图块完成验证",
		"loading":            "加载中...",
		"error":              "出现错误，请重试",
		"protected_by":       "EngageLab CAPTCHA 安全验证",
		"accessibility.alt":  "安全验证挑战",
		"retry":              "重试",
		"close":              "关闭",
	}

	i.translations["ja"] = map[string]string{
		"slider.title":       "スライダーをドラッグして確認",
		"slider.success":     "確認成功",
		"slider.fail":        "もう一度お試しください",
		"click.title":        "順番にクリックしてください",
		"click.success":      "確認成功",
		"click.fail":         "順番が違います。もう一度お試しください",
		"puzzle.title":       "パズルをドラッグして完成させてください",
		"loading":            "読み込み中...",
		"error":              "エラーが発生しました。もう一度お試しください。",
		"protected_by":       "EngageLab CAPTCHAによる保護",
		"retry":              "再試行",
		"close":              "閉じる",
	}

	i.translations["ko"] = map[string]string{
		"slider.title":       "슬라이더를 드래그하여 인증",
		"slider.success":     "인증 성공",
		"slider.fail":        "다시 시도해 주세요",
		"click.title":        "순서대로 클릭하세요",
		"loading":            "로딩 중...",
		"protected_by":       "EngageLab CAPTCHA로 보호됨",
		"retry":              "다시 시도",
	}

	i.translations["es"] = map[string]string{
		"slider.title":       "Arrastra el control para verificar",
		"slider.success":     "Verificacion exitosa",
		"slider.fail":        "Por favor, intente de nuevo",
		"click.title":        "Haz clic en los elementos en orden",
		"loading":            "Cargando...",
		"protected_by":       "Protegido por EngageLab CAPTCHA",
		"retry":              "Reintentar",
	}

	i.translations["pt"] = map[string]string{
		"slider.title":       "Arraste o controle para verificar",
		"slider.success":     "Verificacao bem-sucedida",
		"slider.fail":        "Tente novamente",
		"click.title":        "Clique nos itens em ordem",
		"loading":            "Carregando...",
		"protected_by":       "Protegido por EngageLab CAPTCHA",
		"retry":              "Tentar novamente",
	}

	i.translations["id"] = map[string]string{
		"slider.title":       "Geser slider untuk verifikasi",
		"slider.success":     "Verifikasi berhasil",
		"slider.fail":        "Silakan coba lagi",
		"click.title":        "Klik item sesuai urutan",
		"loading":            "Memuat...",
		"protected_by":       "Dilindungi oleh EngageLab CAPTCHA",
		"retry":              "Coba lagi",
	}

	i.translations["th"] = map[string]string{
		"slider.title":       "ลากตัวเลื่อนเพื่อยืนยัน",
		"slider.success":     "ยืนยันสำเร็จ",
		"slider.fail":        "กรุณาลองอีกครั้ง",
		"click.title":        "คลิกรายการตามลำดับ",
		"loading":            "กำลังโหลด...",
		"protected_by":       "ปกป้องโดย EngageLab CAPTCHA",
		"retry":              "ลองอีกครั้ง",
	}

	i.translations["vi"] = map[string]string{
		"slider.title":       "Keo thanh truot de xac minh",
		"slider.success":     "Xac minh thanh cong",
		"slider.fail":        "Vui long thu lai",
		"click.title":        "Nhan vao cac muc theo thu tu",
		"loading":            "Dang tai...",
		"protected_by":       "Bao ve boi EngageLab CAPTCHA",
		"retry":              "Thu lai",
	}

	i.translations["ar"] = map[string]string{
		"slider.title":       "اسحب شريط التمرير للتحقق",
		"slider.success":     "تم التحقق بنجاح",
		"slider.fail":        "يرجى المحاولة مرة أخرى",
		"click.title":        "انقر على العناصر بالترتيب",
		"loading":            "جاري التحميل...",
		"protected_by":       "محمي بواسطة EngageLab CAPTCHA",
		"retry":              "إعادة المحاولة",
	}
}
