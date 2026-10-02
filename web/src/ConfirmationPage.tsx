import {useEffect, useRef, useState} from "react";
import {useParams, useNavigate} from "react-router-dom";
import {API_URL} from "./App.tsx";
import "./ConfirmationPage.css";

type Status = 'idle' | 'loading' | 'success' | 'invalid' | 'error';

const content: Record<Status, { title: string; text: string }> = {
    idle: {
        title: 'Confirm your account',
        text: "You're one step away. Activate your account to start using GopherSocial.",
    },
    loading: {
        title: 'Confirm your account',
        text: "You're one step away. Activate your account to start using GopherSocial.",
    },
    success: {
        title: 'Your account is active',
        text: 'Welcome to GopherSocial! You can now sign in and start posting.',
    },
    invalid: {
        title: 'This link is no longer valid',
        text: 'It has expired or has already been used. If you already activated your account, you can sign in.',
    },
    error: {
        title: 'Something went wrong',
        text: "We couldn't activate your account right now. Check your connection and try again.",
    },
};

const MailIcon = () => (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <rect x="3" y="5" width="18" height="14" rx="2"/>
        <path d="m4 7 8 6 8-6"/>
    </svg>
);

const CheckIcon = () => (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="m5 12.5 4.5 4.5L19 7.5"/>
    </svg>
);

const AlertIcon = () => (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <circle cx="12" cy="12" r="9"/>
        <path d="M12 7.5v5.5"/>
        <path d="M12 16.5h.01"/>
    </svg>
);

export const ConfirmationPage = () => {
    const {token = ''} = useParams();
    const navigate = useNavigate();
    const [status, setStatus] = useState<Status>(token ? 'idle' : 'invalid');
    const titleRef = useRef<HTMLHeadingElement>(null);
    const isFirstRender = useRef(true);

    // Move focus to the new heading so screen readers announce the result
    useEffect(() => {
        if (isFirstRender.current) {
            isFirstRender.current = false;
            return;
        }
        if (status !== 'loading') {
            titleRef.current?.focus();
        }
    }, [status]);

    const handleConfirm = async () => {
        setStatus('loading');
        try {
            const response = await fetch(`${API_URL}/users/activate/${encodeURIComponent(token)}`, {
                method: 'PUT',
            });

            if (response.ok) {
                setStatus('success');
            } else if (response.status === 404) {
                setStatus('invalid');
            } else {
                setStatus('error');
            }
        } catch {
            // network failure, API down or CORS refusal
            setStatus('error');
        }
    };

    const {title, text} = content[status];
    const isLoading = status === 'loading';

    let icon = <MailIcon/>;
    let iconClass = 'confirm__icon';
    if (status === 'success') {
        icon = <CheckIcon/>;
        iconClass += ' confirm__icon--success';
    } else if (status === 'invalid' || status === 'error') {
        icon = <AlertIcon/>;
        iconClass += ' confirm__icon--error';
    }

    return (
        <main className="confirm">
            <div className="confirm__card">
                <p className="confirm__brand">GopherSocial</p>

                <div className={iconClass}>{icon}</div>

                <div aria-live="polite">
                    <h1 className="confirm__title" ref={titleRef} tabIndex={-1}>{title}</h1>
                    <p className="confirm__text">{text}</p>
                </div>

                {(status === 'idle' || status === 'loading') && (
                    <button
                        type="button"
                        className="confirm__button"
                        onClick={handleConfirm}
                        disabled={isLoading}
                        aria-busy={isLoading}
                    >
                        {isLoading && <span className="confirm__spinner" aria-hidden="true"/>}
                        {isLoading ? 'Activating…' : 'Activate my account'}
                    </button>
                )}

                {status === 'success' && (
                    <button type="button" className="confirm__button" onClick={() => navigate('/')}>
                        Continue to GopherSocial
                    </button>
                )}

                {status === 'invalid' && (
                    <button type="button" className="confirm__button confirm__button--secondary" onClick={() => navigate('/')}>
                        Back to home
                    </button>
                )}

                {status === 'error' && (
                    <button type="button" className="confirm__button" onClick={handleConfirm}>
                        Try again
                    </button>
                )}

                {(status === 'idle' || status === 'loading') && (
                    <p className="confirm__footnote">Didn't create an account? You can safely close this page.</p>
                )}
            </div>
        </main>
    );
};
